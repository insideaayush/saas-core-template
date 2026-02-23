-- Personal workspaces are implemented as organizations with kind='personal'.
-- They are enforced to be single-member and owned by a specific user.

ALTER TABLE organizations
  ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'team',
  ADD COLUMN IF NOT EXISTS personal_owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE organizations
  ADD CONSTRAINT organizations_kind_check CHECK (kind IN ('personal', 'team'));

CREATE INDEX IF NOT EXISTS idx_organizations_personal_owner ON organizations(personal_owner_user_id);

-- Enforce: personal org must have an owner.
CREATE OR REPLACE FUNCTION enforce_personal_org_owner()
RETURNS trigger AS $$
BEGIN
  IF NEW.kind = 'personal' AND NEW.personal_owner_user_id IS NULL THEN
    RAISE EXCEPTION 'personal organization requires personal_owner_user_id';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_enforce_personal_org_owner ON organizations;
CREATE TRIGGER trg_enforce_personal_org_owner
BEFORE INSERT OR UPDATE ON organizations
FOR EACH ROW
EXECUTE FUNCTION enforce_personal_org_owner();

-- Enforce: personal org is single-member and membership must be the owner.
CREATE OR REPLACE FUNCTION enforce_personal_org_membership()
RETURNS trigger AS $$
DECLARE
  org_kind TEXT;
  owner_id UUID;
  other_member_count INT;
BEGIN
  SELECT kind, personal_owner_user_id INTO org_kind, owner_id
  FROM organizations
  WHERE id = NEW.organization_id;

  IF org_kind = 'personal' THEN
    IF owner_id IS NULL THEN
      RAISE EXCEPTION 'personal organization missing owner';
    END IF;

    IF NEW.user_id <> owner_id THEN
      RAISE EXCEPTION 'personal organization membership must be the owner';
    END IF;

    SELECT COUNT(*) INTO other_member_count
    FROM organization_members
    WHERE organization_id = NEW.organization_id
      AND user_id <> owner_id
      AND (TG_OP = 'INSERT' OR id <> NEW.id);

    IF other_member_count > 0 THEN
      RAISE EXCEPTION 'personal organization must be single-member';
    END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_enforce_personal_org_membership ON organization_members;
CREATE TRIGGER trg_enforce_personal_org_membership
BEFORE INSERT OR UPDATE ON organization_members
FOR EACH ROW
EXECUTE FUNCTION enforce_personal_org_membership();

