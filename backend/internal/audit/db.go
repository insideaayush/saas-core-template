package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBRecorder struct {
	db *pgxpool.Pool
}

func NewDBRecorder(db *pgxpool.Pool) *DBRecorder {
	return &DBRecorder{db: db}
}

func (r *DBRecorder) Record(ctx context.Context, event Event) error {
	action := strings.TrimSpace(event.Action)
	if action == "" {
		return fmt.Errorf("missing audit action")
	}

	encoded, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("encode audit data: %w", err)
	}

	var orgID any
	if strings.TrimSpace(event.OrganizationID) != "" {
		orgID = strings.TrimSpace(event.OrganizationID)
	}
	var userID any
	if strings.TrimSpace(event.UserID) != "" {
		userID = strings.TrimSpace(event.UserID)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO audit_events (organization_id, user_id, action, data)
		VALUES ($1::uuid, $2::uuid, $3, $4::jsonb)
	`, orgID, userID, action, string(encoded))
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func (r *DBRecorder) ListByOrganization(ctx context.Context, organizationID string, limit int) ([]EventRecord, error) {
	if strings.TrimSpace(organizationID) == "" {
		return nil, fmt.Errorf("missing organizationID")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.Query(ctx, `
		SELECT id::text,
		       COALESCE(organization_id::text, ''),
		       COALESCE(user_id::text, ''),
		       action,
		       data::text,
		       created_at
		FROM audit_events
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, organizationID, limit)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	records := []EventRecord{}
	for rows.Next() {
		var rec EventRecord
		var dataText string
		var created time.Time
		if err := rows.Scan(&rec.ID, &rec.OrganizationID, &rec.UserID, &rec.Action, &dataText, &created); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		rec.CreatedAt = created.UTC().Format(time.RFC3339)

		var parsed map[string]any
		if err := json.Unmarshal([]byte(dataText), &parsed); err == nil && parsed != nil {
			rec.Data = parsed
		} else {
			rec.Data = map[string]any{}
		}

		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}

	return records, nil
}
