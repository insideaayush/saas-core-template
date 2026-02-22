package files

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UploadURLResponse struct {
	FileID     string            `json:"fileId"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	UploadType string            `json:"uploadType"` // "direct" or "presigned"
}

type Service struct {
	db *pgxpool.Pool

	provider string
	diskPath string

	s3 *S3Provider
}

type Config struct {
	Provider string
	DiskPath string
	S3       *S3Provider
}

func NewService(db *pgxpool.Pool, cfg Config) *Service {
	return &Service{
		db:       db,
		provider: strings.ToLower(strings.TrimSpace(defaultString(cfg.Provider, "disk"))),
		diskPath: strings.TrimSpace(defaultString(cfg.DiskPath, "./.data/uploads")),
		s3:       cfg.S3,
	}
}

type CreateInput struct {
	OrganizationID string
	UploaderUserID string
	Filename       string
	ContentType    string
}

func (s *Service) CreateUploadURL(ctx context.Context, input CreateInput, apiBaseURL string) (UploadURLResponse, error) {
	filename := strings.TrimSpace(input.Filename)
	contentType := strings.TrimSpace(input.ContentType)
	if filename == "" {
		return UploadURLResponse{}, fmt.Errorf("filename is required")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	provider := s.provider
	if provider == "" {
		provider = "disk"
	}

	var id string
	storageKey := buildStorageKey(input.OrganizationID, filename)
	if err := s.db.QueryRow(ctx, `
		INSERT INTO file_objects (organization_id, uploader_user_id, filename, content_type, provider, storage_key, status)
		VALUES ($1::uuid, NULLIF($2, '')::uuid, $3, $4, $5, $6, 'pending')
		RETURNING id::text
	`, input.OrganizationID, input.UploaderUserID, filename, contentType, provider, storageKey).Scan(&id); err != nil {
		return UploadURLResponse{}, fmt.Errorf("insert file object: %w", err)
	}

	switch provider {
	case "disk":
		return UploadURLResponse{
			FileID:     id,
			Method:     http.MethodPost,
			URL:        strings.TrimRight(apiBaseURL, "/") + "/api/v1/files/" + id + "/upload",
			Headers:    map[string]string{},
			UploadType: "direct",
		}, nil
	case "s3":
		if s.s3 == nil {
			return UploadURLResponse{}, fmt.Errorf("s3 not configured")
		}

		url, headers, err := s.s3.PresignPut(ctx, storageKey, contentType, 10*time.Minute)
		if err != nil {
			return UploadURLResponse{}, err
		}

		return UploadURLResponse{
			FileID:     id,
			Method:     http.MethodPut,
			URL:        url,
			Headers:    headers,
			UploadType: "presigned",
		}, nil
	default:
		return UploadURLResponse{}, fmt.Errorf("unknown FILE_STORAGE_PROVIDER %q (expected disk|s3)", provider)
	}
}

func (s *Service) HandleDirectUpload(ctx context.Context, organizationID string, fileID string, file multipart.File, header *multipart.FileHeader) error {
	record, err := s.getFile(ctx, organizationID, fileID)
	if err != nil {
		return err
	}
	if record.Provider != "disk" {
		return fmt.Errorf("direct upload not supported for provider %q", record.Provider)
	}

	if err := os.MkdirAll(s.diskPath, 0o755); err != nil {
		return fmt.Errorf("create upload dir: %w", err)
	}

	targetPath := filepath.Join(s.diskPath, filepath.FromSlash(record.StorageKey))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create upload path: %w", err)
	}

	tmpPath := targetPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, io.LimitReader(file, 25*1024*1024))
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("commit file: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		UPDATE file_objects
		SET status = 'uploaded',
		    size_bytes = $1,
		    updated_at = now()
		WHERE id = $2 AND organization_id = $3
	`, written, fileID, organizationID)
	if err != nil {
		return fmt.Errorf("update file status: %w", err)
	}

	_ = header
	return nil
}

func (s *Service) MarkUploaded(ctx context.Context, organizationID string, fileID string, sizeBytes int64) error {
	// Allow marking s3 uploads as complete after a successful presigned PUT.
	_, err := s.db.Exec(ctx, `
		UPDATE file_objects
		SET status = 'uploaded',
		    size_bytes = NULLIF($1, 0),
		    updated_at = now()
		WHERE id = $2 AND organization_id = $3
	`, sizeBytes, fileID, organizationID)
	if err != nil {
		return fmt.Errorf("mark uploaded: %w", err)
	}
	return nil
}

func (s *Service) GetDownloadURL(ctx context.Context, organizationID string, fileID string, apiBaseURL string) (string, error) {
	record, err := s.getFile(ctx, organizationID, fileID)
	if err != nil {
		return "", err
	}

	switch record.Provider {
	case "disk":
		return strings.TrimRight(apiBaseURL, "/") + "/api/v1/files/" + fileID + "/download", nil
	case "s3":
		if s.s3 == nil {
			return "", fmt.Errorf("s3 not configured")
		}
		url, err := s.s3.PresignGet(ctx, record.StorageKey, 10*time.Minute)
		if err != nil {
			return "", err
		}
		return url, nil
	default:
		return "", fmt.Errorf("unknown provider %q", record.Provider)
	}
}

func (s *Service) ServeDirectDownload(w http.ResponseWriter, r *http.Request, organizationID string, fileID string) error {
	record, err := s.getFile(r.Context(), organizationID, fileID)
	if err != nil {
		return err
	}
	if record.Provider != "disk" {
		return fmt.Errorf("direct download not supported for provider %q", record.Provider)
	}

	path := filepath.Join(s.diskPath, filepath.FromSlash(record.StorageKey))
	w.Header().Set("Content-Type", record.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", record.Filename))
	http.ServeFile(w, r, path)
	return nil
}

type fileRecord struct {
	ID          string
	Provider    string
	StorageKey  string
	Filename    string
	ContentType string
}

func (s *Service) getFile(ctx context.Context, organizationID string, fileID string) (fileRecord, error) {
	var rec fileRecord
	if err := s.db.QueryRow(ctx, `
		SELECT id::text, provider, storage_key, filename, content_type
		FROM file_objects
		WHERE id = $1 AND organization_id = $2
	`, fileID, organizationID).Scan(&rec.ID, &rec.Provider, &rec.StorageKey, &rec.Filename, &rec.ContentType); err != nil {
		return fileRecord{}, fmt.Errorf("file not found")
	}
	rec.Provider = strings.ToLower(strings.TrimSpace(rec.Provider))
	return rec, nil
}

func buildStorageKey(organizationID string, filename string) string {
	// Use a stable prefix per org and keep filename sanitized for readability.
	clean := strings.TrimSpace(filename)
	clean = strings.ReplaceAll(clean, "..", "_")
	clean = strings.ReplaceAll(clean, "/", "_")
	clean = strings.ReplaceAll(clean, "\\", "_")
	if clean == "" {
		clean = "upload.bin"
	}

	return path.Join("org", strings.TrimSpace(organizationID), fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), clean))
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
