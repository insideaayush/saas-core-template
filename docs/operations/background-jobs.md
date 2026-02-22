# Background Jobs

This template includes a minimal Postgres-backed job queue and a separate worker process.

## Local development

1. Start infra: `make infra-up`
2. Run API: `make dev-api`
3. Run worker in another terminal: `make dev-worker`

Jobs are stored in the `jobs` table and claimed with `FOR UPDATE SKIP LOCKED`.

## Configuration

Backend env vars:

- `JOBS_ENABLED=true|false`
- `JOBS_WORKER_ID=<string>` (worker identity)
- `JOBS_POLL_INTERVAL=1s` (poll interval)

## Current job types

- `send_email`: sends a transactional email using the configured email provider.

