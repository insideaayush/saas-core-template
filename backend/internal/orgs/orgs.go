package orgs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"saas-core-template/backend/internal/audit"
	"saas-core-template/backend/internal/jobs"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrInviteAlreadyExists = errors.New("invite already exists")
	ErrInviteAlreadyUsed   = errors.New("invite already used")
	ErrInviteEmailMismatch = errors.New("invite email mismatch")
	ErrInvalidOrganization = errors.New("invalid organization")
)

type Service struct {
	db    *pgxpool.Pool
	jobs  jobs.Enqueuer
	audit audit.Recorder
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Kind string `json:"kind"`
	Role string `json:"role"`
}

type Member struct {
	UserID       string    `json:"userId"`
	PrimaryEmail string    `json:"primaryEmail"`
	Role         string    `json:"role"`
	JoinedAt     time.Time `json:"joinedAt"`
}

type Invite struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organizationId"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Token          string     `json:"token"`
	CreatedAt      time.Time  `json:"createdAt"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
}

type CreateOrgInput struct {
	Name string
	Slug string
}

type CreateInviteInput struct {
	OrganizationID  string
	InvitedByUserID string
	Email           string
	Role            string
}

type AcceptInviteInput struct {
	Token  string
	UserID string
	Email  string
}

type UpdateMemberRoleInput struct {
	OrganizationID string
	UserID         string
	Role           string
}

type RemoveMemberInput struct {
	OrganizationID string
	UserID         string
}

func NewService(db *pgxpool.Pool, opts ...func(*Service)) *Service {
	s := &Service{db: db, audit: audit.NewNoop()}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
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

func (s *Service) ListForUser(ctx context.Context, userID string) ([]Organization, error) {
	rows, err := s.db.Query(ctx, `
		SELECT o.id::text, o.name, o.slug, o.kind, om.role
		FROM organizations o
		INNER JOIN organization_members om ON om.organization_id = o.id
		WHERE om.user_id = $1
		ORDER BY om.created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var out []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.Kind, &org.Role); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		out = append(out, org)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list organizations rows: %w", rows.Err())
	}
	return out, nil
}

func (s *Service) CreateTeamOrganization(ctx context.Context, creatorUserID string, input CreateOrgInput) (Organization, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Organization{}, fmt.Errorf("missing name")
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		slug = slugify(name)
	}
	if slug == "" {
		return Organization{}, fmt.Errorf("invalid slug")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var org Organization
	org.Name = name
	org.Slug = slug
	org.Kind = "team"
	org.Role = "owner"

	baseSlug := slug
	for i := 0; i < 5; i++ {
		err := tx.QueryRow(ctx, `
			INSERT INTO organizations (name, slug, kind)
			VALUES ($1, $2, 'team')
			RETURNING id::text
		`, name, slug).Scan(&org.ID)
		if err == nil {
			break
		}
		if isUniqueViolation(err) {
			slug = fmt.Sprintf("%s-%s", baseSlug, randomSuffix())
			continue
		}
		return Organization{}, fmt.Errorf("insert organization: %w", err)
	}
	if org.ID == "" {
		return Organization{}, fmt.Errorf("failed to allocate unique slug")
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, org.ID, creatorUserID); err != nil {
		return Organization{}, fmt.Errorf("insert organization member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Organization{}, fmt.Errorf("commit create org: %w", err)
	}

	_ = s.audit.Record(ctx, audit.Event{
		OrganizationID: org.ID,
		UserID:         creatorUserID,
		Action:         "organization_created",
		Data:           map[string]any{"kind": "team"},
	})

	return org, nil
}

func (s *Service) ListMembers(ctx context.Context, organizationID string) ([]Member, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.id::text, COALESCE(u.primary_email, ''), om.role, om.created_at
		FROM organization_members om
		INNER JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		ORDER BY om.created_at ASC
	`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.PrimaryEmail, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		out = append(out, m)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list members rows: %w", rows.Err())
	}
	return out, nil
}

func (s *Service) CreateInvite(ctx context.Context, input CreateInviteInput) (Invite, error) {
	email := normalizeEmail(input.Email)
	if email == "" {
		return Invite{}, fmt.Errorf("missing email")
	}

	role := strings.ToLower(strings.TrimSpace(input.Role))
	if role == "" {
		role = "member"
	}
	if role != "member" && role != "admin" {
		return Invite{}, fmt.Errorf("invalid role")
	}

	var orgKind string
	if err := s.db.QueryRow(ctx, `SELECT kind FROM organizations WHERE id = $1`, input.OrganizationID).Scan(&orgKind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invite{}, ErrNotFound
		}
		return Invite{}, fmt.Errorf("load organization kind: %w", err)
	}
	if orgKind != "team" {
		return Invite{}, ErrInvalidOrganization
	}

	token, err := newToken(16)
	if err != nil {
		return Invite{}, fmt.Errorf("generate token: %w", err)
	}

	var invite Invite
	err = s.db.QueryRow(ctx, `
		INSERT INTO organization_invites (organization_id, email, role, token, invited_by_user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, organization_id::text, email, role, token, created_at, accepted_at
	`, input.OrganizationID, email, role, token, input.InvitedByUserID).Scan(
		&invite.ID,
		&invite.OrganizationID,
		&invite.Email,
		&invite.Role,
		&invite.Token,
		&invite.CreatedAt,
		&invite.AcceptedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return Invite{}, ErrInviteAlreadyExists
		}
		return Invite{}, fmt.Errorf("insert invite: %w", err)
	}

	_ = s.audit.Record(ctx, audit.Event{
		OrganizationID: input.OrganizationID,
		UserID:         input.InvitedByUserID,
		Action:         "organization_invite_created",
		Data:           map[string]any{"email": email, "role": role},
	})

	return invite, nil
}

func (s *Service) EnqueueInviteEmail(ctx context.Context, invite Invite, acceptURL string) error {
	if s.jobs == nil {
		return nil
	}
	if strings.TrimSpace(invite.Email) == "" || strings.TrimSpace(acceptURL) == "" {
		return nil
	}

	orgName := "your workspace"
	_ = s.db.QueryRow(ctx, `SELECT name FROM organizations WHERE id = $1`, invite.OrganizationID).Scan(&orgName)

	subject := fmt.Sprintf("You're invited to join %s", orgName)
	text := fmt.Sprintf("You have been invited to join %s.\n\nAccept: %s\n", orgName, acceptURL)

	_, err := s.jobs.Enqueue(ctx, "send_email", map[string]any{
		"to":      invite.Email,
		"subject": subject,
		"text":    text,
	}, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("enqueue invite email: %w", err)
	}
	return nil
}

func (s *Service) AcceptInvite(ctx context.Context, input AcceptInviteInput) (Organization, error) {
	token := strings.TrimSpace(input.Token)
	if token == "" {
		return Organization{}, fmt.Errorf("missing token")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var inviteID string
	var orgID string
	var email string
	var role string
	var acceptedAt *time.Time
	if err := tx.QueryRow(ctx, `
		SELECT id::text, organization_id::text, email, role, accepted_at
		FROM organization_invites
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&inviteID, &orgID, &email, &role, &acceptedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}
		return Organization{}, fmt.Errorf("load invite: %w", err)
	}

	if acceptedAt != nil {
		return Organization{}, ErrInviteAlreadyUsed
	}

	if normalizeEmail(input.Email) == "" || normalizeEmail(input.Email) != normalizeEmail(email) {
		return Organization{}, ErrInviteEmailMismatch
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role, updated_at = now()
	`, orgID, input.UserID, role); err != nil {
		return Organization{}, fmt.Errorf("insert membership: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE organization_invites
		SET accepted_at = now(),
		    accepted_by_user_id = $1,
		    updated_at = now()
		WHERE id = $2
	`, input.UserID, inviteID); err != nil {
		return Organization{}, fmt.Errorf("mark invite accepted: %w", err)
	}

	var org Organization
	if err := tx.QueryRow(ctx, `
		SELECT o.id::text, o.name, o.slug, o.kind, om.role
		FROM organizations o
		INNER JOIN organization_members om ON om.organization_id = o.id
		WHERE o.id = $1 AND om.user_id = $2
	`, orgID, input.UserID).Scan(&org.ID, &org.Name, &org.Slug, &org.Kind, &org.Role); err != nil {
		return Organization{}, fmt.Errorf("load organization: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Organization{}, fmt.Errorf("commit accept invite: %w", err)
	}

	_ = s.audit.Record(ctx, audit.Event{
		OrganizationID: orgID,
		UserID:         input.UserID,
		Action:         "organization_invite_accepted",
		Data:           map[string]any{"email": normalizeEmail(email), "role": role},
	})

	return org, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) error {
	role := strings.ToLower(strings.TrimSpace(input.Role))
	if role != "owner" && role != "admin" && role != "member" {
		return fmt.Errorf("invalid role")
	}

	var orgKind string
	if err := s.db.QueryRow(ctx, `SELECT kind FROM organizations WHERE id = $1`, input.OrganizationID).Scan(&orgKind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("load organization kind: %w", err)
	}
	if orgKind != "team" {
		return ErrInvalidOrganization
	}

	ct, err := s.db.Exec(ctx, `
		UPDATE organization_members
		SET role = $1, updated_at = now()
		WHERE organization_id = $2 AND user_id = $3
	`, role, input.OrganizationID, input.UserID)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	var orgKind string
	if err := s.db.QueryRow(ctx, `SELECT kind FROM organizations WHERE id = $1`, input.OrganizationID).Scan(&orgKind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("load organization kind: %w", err)
	}
	if orgKind != "team" {
		return ErrInvalidOrganization
	}

	ct, err := s.db.Exec(ctx, `
		DELETE FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`, input.OrganizationID, input.UserID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func newToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

var slugAllowed = regexp.MustCompile(`[^a-z0-9-]+`)

func slugify(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return ""
	}

	trimmed = strings.ReplaceAll(trimmed, "_", "-")
	trimmed = strings.ReplaceAll(trimmed, " ", "-")
	trimmed = slugAllowed.ReplaceAllString(trimmed, "")
	trimmed = strings.Trim(trimmed, "-")
	for strings.Contains(trimmed, "--") {
		trimmed = strings.ReplaceAll(trimmed, "--", "-")
	}
	return trimmed
}

func randomSuffix() string {
	token, err := newToken(3)
	if err != nil {
		return "alt"
	}
	return token
}

func normalizeEmail(email string) string {
	trimmed := strings.TrimSpace(strings.ToLower(email))
	if strings.Contains(trimmed, " ") {
		return ""
	}
	if !strings.Contains(trimmed, "@") {
		return ""
	}
	return trimmed
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value") || strings.Contains(msg, "unique constraint")
}
