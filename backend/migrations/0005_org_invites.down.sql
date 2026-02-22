DROP INDEX IF EXISTS uq_organization_invites_active;
DROP INDEX IF EXISTS idx_organization_invites_org_id;

ALTER TABLE organization_invites
  DROP CONSTRAINT IF EXISTS organization_invites_role_check;

DROP TABLE IF EXISTS organization_invites;

