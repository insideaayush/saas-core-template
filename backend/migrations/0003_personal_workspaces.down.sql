DROP TRIGGER IF EXISTS trg_enforce_personal_org_membership ON organization_members;
DROP FUNCTION IF EXISTS enforce_personal_org_membership;

DROP TRIGGER IF EXISTS trg_enforce_personal_org_owner ON organizations;
DROP FUNCTION IF EXISTS enforce_personal_org_owner;

DROP INDEX IF EXISTS idx_organizations_personal_owner;

ALTER TABLE organizations DROP CONSTRAINT IF EXISTS organizations_kind_check;

ALTER TABLE organizations
  DROP COLUMN IF EXISTS personal_owner_user_id,
  DROP COLUMN IF EXISTS kind;

