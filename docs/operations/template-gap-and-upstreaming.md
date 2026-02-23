# Template Gaps and Upstreaming Guide

This document serves two purposes:

1) Track what is still missing in `saas-core-template`.
2) Provide instructions for agents working in *child projects* (projects created from this template) on how to upstream improvements back into the template when appropriate.

## Current status (what works today)

- Local infra via Docker Compose (`postgres`, `redis`, `otel-collector`).
- Backend API + worker in Go; Postgres-backed jobs, email adapter, file uploads, audit logs.
- App-owned organizations with personal workspace + team orgs, invites, and RBAC (member/admin/owner).
- Frontend on Next.js (Vercel target) with shadcn/ui baseline; minimal org switcher + invite acceptance flow.
- Local smoke test script: `make smoke-local` (use `SMOKE_ARGS=--skip-ui` if your Node version can’t run Next.js).
- Migration runner: `make migrate-up` (tracks applied migrations in `schema_migrations`).

## Remaining work (template backlog)

### Local E2E (developer experience)

- Full smoke test including UI (`make smoke-local` without `--skip-ui`) requires Node 20+ on the developer machine.
- Add a “prod smoke” runbook/script to verify Vercel + Render wiring and provider integrations end-to-end.

### Security hardening

- Replace permissive CORS (`*`) with an allowlist for production.
- Add rate limiting / abuse protections for public endpoints.
- Add request size limits consistently (some exist implicitly, but not centrally enforced).

### RBAC and authorization coverage

- Keep expanding RBAC coverage as new endpoints are added (deny-by-default, role-gated mutations, sensitive reads).
- Add dedicated org settings/member management pages in the UI with clear role gating.

### Production migration automation

- Decide where `backend/cmd/migrate up` runs in production (manual, one-off Render job, or deployment step) and document the exact procedure.

## When a child project should upstream a change

Upstream changes when they are **template-shaped**:

- Cross-cutting infrastructure (auth boundaries, tenancy, billing wiring, observability, migrations, background jobs).
- Generic product modules that most SaaS apps need (org management, audit logging, file uploads, email, analytics/error reporting/support widgets).
- Developer-experience improvements (smoke tests, scripts, docs, safer defaults).
- Provider portability improvements (adapters/interfaces, migration playbooks).

Do **not** upstream:

- Business logic, domain models, or product-specific UI flows.
- Customer-specific integrations, hardcoded pricing, or bespoke schemas.
- Anything requiring paid providers by default (managed providers must remain optional and swappable).

## Upstreaming workflow (agent instructions for child projects)

### 1) Triage: is this template-worthy?

Use this quick filter:

- Would I want this in my *next* SaaS experiment?
- Can it ship with “console/noop” defaults and work locally without paid accounts?
- Does it keep provider-specific types out of domain models?
- Does it preserve tenant scoping and deny-by-default authorization?

If “yes” to all, it’s a good upstream candidate.

### 2) Keep changes portable

When implementing in the child project:

- Put provider-specific code behind interfaces/adapters.
- Add feature toggles / env-driven provider selection (for example `*_PROVIDER=console|managed|none`).
- Keep migrations additive and safe; avoid destructive schema changes unless necessary.
- Avoid coupling to the child project name (do not rename module paths back in the template).

### 3) Package the upstream change

Before upstreaming:

- Extract the generic portion into a clean commit (or a short sequence of commits) with no business logic.
- Add/update docs in the template:
  - Architecture changes → `docs/architecture/*`
  - Ops/runbooks → `docs/operations/*`
  - Any new env vars → `.env.example`, `backend/.env.example`, `frontend/.env.example`
- Add/extend smoke coverage where possible:
  - local: `scripts/smoke-local.sh`
  - migrations: `backend/cmd/migrate`
- Run validations (or ensure CI will cover them).

### 4) Upstream mechanics

Recommended approach:

1. In the child repo, create a branch containing only template-worthy commits.
2. Apply the same commits to the template repo:
   - either by `git cherry-pick` onto a `feature/*` branch in the template repo,
   - or by opening a PR from the child repo’s branch (if you maintain a remote that can target the template).
3. In the template repo:
   - ensure no secrets are added
   - ensure docs are updated
   - ensure local smoke still passes (use `SMOKE_ARGS=--skip-ui` if needed)

## Notes for template initialization

Each child project should run the initialization protocol once:

- `./scripts/init-template.sh "<new-project-name>"`
- Verify rename integrity: `rg "saas-core-template|saas_core_template"`

This template is intentionally “copy forward”; child projects are not guaranteed downstream updates. The upstreaming process above is the mechanism for voluntarily contributing generic improvements back to the template for future projects.

