# Agent Workflow Runbook

This runbook defines the standard operating flow for AI-agent-assisted development in this repository.

## 1) Initialization and context

1. Confirm template naming has been initialized:
   - `./scripts/init-template.sh "<new-project-name>"` (one-time per new clone).
2. Verify no stale template identifiers remain:
   - `rg "saas-core-template|saas_core_template"`
3. Read guardrails:
   - `AGENTS.md`
   - Relevant docs in `docs/architecture/` and `docs/operations/`.

## 2) Planning and scope

- Keep scope tightly aligned to the requested task.
- Prefer incremental changes with clear interfaces.
- Avoid opportunistic refactors unless explicitly requested.

## 3) Implementation rules

- Keep auth and billing provider SDK calls behind adapter/services.
- Preserve internal domain ownership (`user_id`, org membership, subscription state).
- Enforce tenant scoping on all tenant-owned data operations.
- Keep webhook handlers idempotent and replay-safe.

## 4) Validation checklist

- Backend:
  - `gofmt` on changed files
  - `go test ./...`
- Frontend:
  - `npm run lint`
  - `npm run typecheck`
  - `npm run build` for route/config changes
- Configuration:
  - Validate env example files are still consistent and complete.
  - Validate deployment config changes reflect new variables (for example `render.yaml` for Render backend, and Vercel project env vars for frontend).
  - Confirm managed integrations are optional and local E2E still works with console/noop defaults (OpenTelemetry, analytics, error reporting, support widget).
  - Confirm no secrets were committed while adding integration variables.

## 5) Documentation and traceability

- Update docs in the same task for any contract or behavior changes.
- Document migration implications when introducing provider-specific behavior.
- Keep onboarding docs synchronized with setup changes.

## 6) Commit and push hygiene

- Never commit secrets or `.env` files.
- Keep commits cohesive and task-specific.
- If push fails, report the exact error and remediation steps.
- Follow branch policy:
  - feature work on `dev` or `feature/*`
  - integrate via `develop`
  - release via `main`
- Keep `VERSION` in SemVer format and update it intentionally for release scope.

## 7) Completion criteria

A task is considered complete when:

- Code changes are implemented and validated.
- Relevant docs are updated.
- No stale template identifiers remain unintentionally.
- Working tree is in the expected post-task state.
