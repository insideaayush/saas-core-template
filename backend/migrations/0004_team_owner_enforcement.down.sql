DROP TRIGGER IF EXISTS trg_enforce_team_has_owner ON organization_members;
DROP FUNCTION IF EXISTS enforce_team_has_owner;

ALTER TABLE organization_members
  DROP CONSTRAINT IF EXISTS organization_members_role_check;

