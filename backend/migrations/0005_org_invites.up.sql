-- Organization invites for team workspaces (email-based).

CREATE TABLE IF NOT EXISTS organization_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    token TEXT NOT NULL UNIQUE,
    invited_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at TIMESTAMPTZ,
    accepted_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL
);

ALTER TABLE organization_invites
  ADD CONSTRAINT organization_invites_role_check CHECK (role IN ('admin', 'member'));

CREATE INDEX IF NOT EXISTS idx_organization_invites_org_id ON organization_invites(organization_id);

-- One outstanding invite per org+email (case-insensitive).
CREATE UNIQUE INDEX IF NOT EXISTS uq_organization_invites_active
ON organization_invites(organization_id, lower(email))
WHERE accepted_at IS NULL;
