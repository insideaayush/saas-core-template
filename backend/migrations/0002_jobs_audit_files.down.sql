DROP INDEX IF EXISTS idx_file_objects_org_created_at;
DROP TABLE IF EXISTS file_objects;

DROP INDEX IF EXISTS idx_audit_events_user_created_at;
DROP INDEX IF EXISTS idx_audit_events_org_created_at;
DROP TABLE IF EXISTS audit_events;

DROP INDEX IF EXISTS idx_jobs_locked_until;
DROP INDEX IF EXISTS idx_jobs_status_run_at;
DROP TABLE IF EXISTS jobs;

