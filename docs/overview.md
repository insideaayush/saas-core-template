# Project Overview and Code Reading Order

This doc is a “start here” guide for understanding the template quickly.

## Reading order (recommended)

1. Guardrails and workflow
   - `AGENTS.md`
   - `docs/operations/agent-workflow.md`
2. Architecture playbooks (what the template is optimizing for)
   - `docs/architecture/auth-and-identity.md`
   - `docs/architecture/multi-tenant-model.md`
   - `docs/architecture/billing-and-pricing.md`
3. Local run + production wiring
   - `README.md`
   - `docs/operations/production-setup-checklist.md`
4. Integrations (optional, managed-first)
   - `docs/operations/observability.md`
   - `docs/operations/product-analytics.md`
   - `docs/operations/error-reporting.md`
   - `docs/operations/support.md`
5. Frontend UI conventions
   - `docs/frontend/ui-design-guide.md`

## Repository layout (what lives where)

- `backend/`
  - `cmd/api/main.go`: composition root (config, DB/Redis, providers, server start).
  - `internal/api/`: HTTP transport (routes, middleware, JSON responses).
  - `internal/auth/`: auth provider adapter + identity mapping + org resolution.
  - `internal/billing/`: billing provider adapter + webhook handling + subscription state.
  - `internal/analytics/`: backend analytics adapter boundary (console/PostHog/noop).
  - `internal/telemetry/`: OpenTelemetry init and exporter selection.
  - `internal/errorreporting/`: backend error reporting adapter (console/Sentry/noop).
  - `migrations/`: SQL migrations (identity, tenancy, billing tables).
- `frontend/`
  - `app/`: Next.js routes and UI shells.
  - `lib/api.ts`: frontend API client to the Go backend.
  - `lib/integrations/*`: frontend integrations (analytics/support/error reporting).
  - `lib/i18n/*`: minimal i18n layer (cookie locale + message catalog).
- `docker-compose.yml`: local Postgres + Redis + local OTel collector.
- `render.yaml`: Render blueprint (backend + Postgres).

## Key request flows

### Frontend → Backend connectivity

- Frontend calls `GET /api/v1/meta` and `GET /api/v1/auth/me` using `NEXT_PUBLIC_API_URL`.

### Auth flow (Clerk)

- Frontend gets a Clerk session token and calls backend with `Authorization: Bearer <token>`.
- Backend verifies token with Clerk, then ensures:
  - internal `users` row exists
  - identity mapping exists in `auth_identities`
  - user has at least one organization membership

### Tenancy / org context

- Frontend sends `X-Organization-ID` (when available).
- Backend resolves org membership and denies by default.

### Billing flow (Stripe)

- Frontend calls checkout/portal endpoints (org-scoped).
- Backend uses internal plan mapping (`plans`) and creates Stripe sessions.
- Stripe sends webhooks to `/api/v1/billing/webhook` which upserts internal subscription state.

## Local-first defaults (important)

- If provider keys are unset:
  - auth and billing endpoints return “not configured” errors
  - telemetry/analytics/error reporting default to console output or no-ops
  - support widget is disabled by default

