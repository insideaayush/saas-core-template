package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"saas-core-template/backend/internal/audit"
	"saas-core-template/backend/internal/jobs"
)

var (
	ErrMissingBearerToken = errors.New("missing bearer token")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNoOrganization     = errors.New("no organization found for user")
)

type VerifiedPrincipal struct {
	Provider       string
	ProviderUserID string
	PrimaryEmail   string
	EmailVerified  bool
}

type Provider interface {
	VerifyToken(ctx context.Context, token string) (VerifiedPrincipal, error)
}

type Service struct {
	provider Provider
	db       *pgxpool.Pool
	jobs     jobs.Enqueuer
	audit    audit.Recorder
}

type User struct {
	ID           string `json:"id"`
	PrimaryEmail string `json:"primaryEmail"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}

func NewService(provider Provider, db *pgxpool.Pool, opts ...func(*Service)) *Service {
	svc := &Service{
		provider: provider,
		db:       db,
		jobs:     nil,
		audit:    audit.NewNoop(),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(svc)
		}
	}

	return svc
}

func WithJobs(enqueuer jobs.Enqueuer) func(*Service) {
	return func(s *Service) {
		s.jobs = enqueuer
	}
}

func WithAudit(recorder audit.Recorder) func(*Service) {
	return func(s *Service) {
		if recorder != nil {
			s.audit = recorder
		}
	}
}

func ExtractBearerToken(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", ErrMissingBearerToken
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", ErrMissingBearerToken
	}

	return strings.TrimSpace(parts[1]), nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	principal, err := s.provider.VerifyToken(ctx, token)
	if err != nil {
		return User{}, fmt.Errorf("verify token: %w", err)
	}

	userID, created, err := s.ensureUserIdentity(ctx, principal)
	if err != nil {
		return User{}, err
	}

	var user User
	if err := s.db.QueryRow(ctx, `SELECT id::text, primary_email FROM users WHERE id = $1`, userID).Scan(&user.ID, &user.PrimaryEmail); err != nil {
		return User{}, fmt.Errorf("load user: %w", err)
	}

	if err := s.ensureDefaultOrganizationForUser(ctx, user); err != nil {
		return User{}, err
	}

	if created {
		_ = s.audit.Record(ctx, audit.Event{
			UserID: user.ID,
			Action: "user_created",
			Data:   map[string]any{"primary_email": user.PrimaryEmail, "provider": principal.Provider},
		})

		if s.jobs != nil && strings.TrimSpace(user.PrimaryEmail) != "" {
			_, _ = s.jobs.Enqueue(ctx, "send_email", map[string]any{
				"kind":    "welcome",
				"to":      user.PrimaryEmail,
				"subject": "Welcome",
				"text":    "Welcome to the app. You're set up and ready to go.",
			}, time.Now().UTC())
		}
	}

	return user, nil
}

func (s *Service) ensureUserIdentity(ctx context.Context, principal VerifiedPrincipal) (string, bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID string
	err = tx.QueryRow(ctx, `
		SELECT user_id::text
		FROM auth_identities
		WHERE provider = $1 AND provider_user_id = $2
	`, principal.Provider, principal.ProviderUserID).Scan(&userID)
	if err == nil {
		if _, err := tx.Exec(ctx, `
			UPDATE auth_identities
			SET provider_email = COALESCE($1, provider_email),
			    email_verified_at = CASE WHEN $2 THEN now() ELSE email_verified_at END,
			    updated_at = now()
			WHERE provider = $3 AND provider_user_id = $4
			`, emptyToNil(principal.PrimaryEmail), principal.EmailVerified, principal.Provider, principal.ProviderUserID); err != nil {
			return "", false, fmt.Errorf("update identity: %w", err)
		}

		if principal.PrimaryEmail != "" {
			if _, err := tx.Exec(ctx, `UPDATE users SET primary_email = $1, updated_at = now() WHERE id = $2`, principal.PrimaryEmail, userID); err != nil {
				return "", false, fmt.Errorf("update user email: %w", err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return "", false, fmt.Errorf("commit existing identity: %w", err)
		}

		return userID, false, nil
	}

	// Create new user and identity mapping when no existing identity is found.
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (primary_email)
		VALUES ($1)
		RETURNING id::text
	`, emptyToNil(principal.PrimaryEmail)).Scan(&userID); err != nil {
		return "", false, fmt.Errorf("insert user: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO auth_identities (user_id, provider, provider_user_id, provider_email, email_verified_at)
		VALUES ($1, $2, $3, $4, CASE WHEN $5 THEN now() ELSE NULL END)
	`, userID, principal.Provider, principal.ProviderUserID, emptyToNil(principal.PrimaryEmail), principal.EmailVerified); err != nil {
		return "", false, fmt.Errorf("insert identity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", false, fmt.Errorf("commit new identity: %w", err)
	}

	return userID, true, nil
}

func (s *Service) ResolveOrganization(ctx context.Context, userID string, requestedOrgID string) (Organization, error) {
	query := `
		SELECT o.id::text, o.name, o.slug, om.role
		FROM organizations o
		INNER JOIN organization_members om ON om.organization_id = o.id
		WHERE om.user_id = $1
	`
	args := []any{userID}

	if requestedOrgID != "" {
		query += ` AND o.id::text = $2`
		args = append(args, requestedOrgID)
	}

	query += ` ORDER BY om.created_at ASC LIMIT 1`

	var org Organization
	if err := s.db.QueryRow(ctx, query, args...).Scan(&org.ID, &org.Name, &org.Slug, &org.Role); err != nil {
		return Organization{}, ErrNoOrganization
	}

	return org, nil
}

func emptyToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return strings.TrimSpace(value)
}

func (s *Service) ensureDefaultOrganizationForUser(ctx context.Context, user User) error {
	var membershipCount int
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM organization_members
		WHERE user_id = $1
	`, user.ID).Scan(&membershipCount); err != nil {
		return fmt.Errorf("count user memberships: %w", err)
	}

	if membershipCount > 0 {
		return nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin default org tx: %w", err)
	}
	defer tx.Rollback(ctx)

	name := defaultOrganizationName(user.PrimaryEmail)
	slug := fmt.Sprintf("workspace-%s", shortKey(user.ID))
	var orgID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, kind, personal_owner_user_id)
		VALUES ($1, $2, 'personal', $3::uuid)
		RETURNING id::text
	`, name, slug, user.ID).Scan(&orgID); err != nil {
		return fmt.Errorf("create default organization: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT (organization_id, user_id) DO NOTHING
	`, orgID, user.ID); err != nil {
		return fmt.Errorf("create organization membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit default organization: %w", err)
	}

	return nil
}

func defaultOrganizationName(primaryEmail string) string {
	if strings.TrimSpace(primaryEmail) == "" {
		return "My Workspace"
	}

	prefix := strings.SplitN(primaryEmail, "@", 2)[0]
	if strings.TrimSpace(prefix) == "" {
		return "My Workspace"
	}

	return fmt.Sprintf("%s's Workspace", prefix)
}

func shortKey(id string) string {
	trimmed := strings.ReplaceAll(strings.TrimSpace(id), "-", "")
	if len(trimmed) >= 8 {
		return trimmed[:8]
	}
	if trimmed == "" {
		return "default"
	}
	return trimmed
}
