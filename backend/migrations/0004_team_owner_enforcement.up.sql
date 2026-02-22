-- Enforce role values and ensure team orgs always retain at least one owner.

ALTER TABLE organization_members
  ADD CONSTRAINT organization_members_role_check CHECK (role IN ('owner', 'admin', 'member'));

CREATE OR REPLACE FUNCTION enforce_team_has_owner()
RETURNS trigger AS $$
DECLARE
  org_kind TEXT;
  owner_count INT;
  target_org_id UUID;
BEGIN
  target_org_id := COALESCE(OLD.organization_id, NEW.organization_id);

  SELECT kind INTO org_kind
  FROM organizations
  WHERE id = target_org_id;

  IF org_kind = 'team' THEN
    -- If the operation would remove or demote an owner, ensure at least one owner remains.
    IF TG_OP = 'DELETE' THEN
      IF OLD.role = 'owner' THEN
        SELECT COUNT(*) INTO owner_count
        FROM organization_members
        WHERE organization_id = OLD.organization_id
          AND role = 'owner'
          AND id <> OLD.id;

        IF owner_count = 0 THEN
          RAISE EXCEPTION 'team organization must have at least one owner';
        END IF;
      END IF;
    ELSIF TG_OP = 'UPDATE' THEN
      IF OLD.role = 'owner' AND NEW.role <> 'owner' THEN
        SELECT COUNT(*) INTO owner_count
        FROM organization_members
        WHERE organization_id = NEW.organization_id
          AND role = 'owner'
          AND id <> NEW.id;

        IF owner_count = 0 THEN
          RAISE EXCEPTION 'team organization must have at least one owner';
        END IF;
      END IF;
    END IF;
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_enforce_team_has_owner ON organization_members;
CREATE TRIGGER trg_enforce_team_has_owner
BEFORE UPDATE OR DELETE ON organization_members
FOR EACH ROW
EXECUTE FUNCTION enforce_team_has_owner();

