# Production Setup Checklist

This checklist wires the template end-to-end in production with:

- Frontend on Vercel
- Backend + Worker + Postgres on Render
- Redis on Upstash
- Providers: Clerk (auth), Stripe (billing), Resend (email), PostHog (analytics), Sentry (error reporting), Crisp (support), Grafana Cloud (telemetry)

Local development should continue to work with console/noop defaults and Docker Compose infra.

## 0) Create provider accounts (once)

- Clerk: create an application.
- Stripe: create an account and products/prices.
- Resend: create an account, verify domain/sender, create API key.
- Upstash: create a Redis database.
- PostHog: create a project.
- Sentry: create a project (frontend + backend can share or be separate).
- Crisp: create a website.
- Grafana Cloud: create a stack with OTLP endpoint + API token.

## 1) Deploy backend + worker + Postgres (Render)

1. In Render, create services from `render.yaml`.
2. Confirm backend service is reachable (Render service URL):
   - `GET /healthz`
   - `GET /readyz` (will fail until DB + Redis configured)
   - `GET /api/v1/meta`

### Backend env vars (Render)

Set these in the backend service:

- Core
  - `APP_ENV=production`
  - `APP_VERSION=<git sha or release>`
  - `APP_BASE_URL=<vercel frontend url>` (used for Stripe return URLs)
  - `DATABASE_URL` (from Render Postgres)
  - `REDIS_URL` (from Upstash)
- Auth (Clerk)
  - `CLERK_SECRET_KEY`
  - `CLERK_API_URL=https://api.clerk.com`
- Billing (Stripe)
  - `STRIPE_SECRET_KEY`
  - `STRIPE_WEBHOOK_SECRET`
  - `STRIPE_PRICE_PRO_MONTHLY`
  - `STRIPE_PRICE_TEAM_MONTHLY`
- Telemetry (Grafana Cloud via OTLP)
  - `OTEL_TRACES_EXPORTER=otlp`
  - `OTEL_SERVICE_NAME=saas-core-template-backend` (or your service name)
  - `OTEL_EXPORTER_OTLP_ENDPOINT=<grafana cloud otlp endpoint>`
  - `OTEL_EXPORTER_OTLP_HEADERS=Authorization=Basic <base64(instance_id:api_token)>`
- Error reporting (Sentry)
  - `ERROR_REPORTING_PROVIDER=sentry`
  - `SENTRY_DSN`
  - `SENTRY_ENVIRONMENT=production`
- Analytics (PostHog)
  - `ANALYTICS_PROVIDER=posthog`
  - `POSTHOG_PROJECT_KEY`
  - `POSTHOG_HOST=https://app.posthog.com` (or your host)
- File uploads (S3 / R2)
  - `FILE_STORAGE_PROVIDER=s3`
  - `S3_BUCKET`
  - `S3_REGION` (R2 uses `auto`)
  - `S3_ENDPOINT` (R2 required)
  - `S3_ACCESS_KEY_ID`
  - `S3_SECRET_ACCESS_KEY`
  - `S3_FORCE_PATH_STYLE=true`

### Worker env vars (Render)

The worker service runs background jobs (emails, future async tasks). Configure:

- Jobs
  - `JOBS_ENABLED=true`
  - `JOBS_WORKER_ID=render`
  - `JOBS_POLL_INTERVAL=1s`
- Email (Resend)
  - `EMAIL_PROVIDER=resend`
  - `EMAIL_FROM=<verified sender>`
  - `RESEND_API_KEY=<api key>`
- Observability / error reporting
  - `OTEL_*` and `SENTRY_*` as above

### Database migration (Render Postgres)

Apply migrations against Render Postgres before using auth/billing/files endpoints.

Recommended (tracks applied migrations in `schema_migrations`):

```bash
cd backend
DATABASE_URL="<render postgres url>" go run ./cmd/migrate up -dir ./migrations
```

Migrations (applied in order):

- `backend/migrations/0001_identity_tenancy_billing.up.sql`
- `backend/migrations/0002_jobs_audit_files.up.sql`
- `backend/migrations/0003_personal_workspaces.up.sql`
- `backend/migrations/0004_team_owner_enforcement.up.sql`
- `backend/migrations/0005_org_invites.up.sql`

## 2) Deploy frontend (Vercel)

1. Import the repo in Vercel.
2. Set project root directory to `frontend/`.
3. Set environment variables.
4. Deploy.

### Frontend env vars (Vercel)

- API
  - `NEXT_PUBLIC_API_URL=<render backend url>`
- Auth (Clerk)
  - `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
- Analytics (PostHog)
  - `NEXT_PUBLIC_ANALYTICS_PROVIDER=posthog`
  - `NEXT_PUBLIC_POSTHOG_KEY`
  - `NEXT_PUBLIC_POSTHOG_HOST=https://app.posthog.com` (or your host)
- Support (Crisp)
  - `NEXT_PUBLIC_SUPPORT_PROVIDER=crisp`
  - `NEXT_PUBLIC_CRISP_WEBSITE_ID`
- Error reporting (Sentry)
  - `NEXT_PUBLIC_ERROR_REPORTING_PROVIDER=sentry`
  - `NEXT_PUBLIC_SENTRY_DSN`
  - `NEXT_PUBLIC_SENTRY_ENVIRONMENT=production`
  - `NEXT_PUBLIC_APP_VERSION=<git sha or release>` (optional)

## 3) Provider dashboard configuration

### Clerk

- Add the Vercel frontend URL to allowed origins / redirect URLs.

### Stripe

- Create price IDs for `pro` and `team` and set them in backend env.
- Configure webhook endpoint:
  - URL: `https://<render-backend>/api/v1/billing/webhook`
  - Events: `checkout.session.completed`, `customer.subscription.created`, `customer.subscription.updated`, `customer.subscription.deleted`

### PostHog

- Ensure project keys match frontend/back env vars.

### Sentry

- Ensure DSNs match frontend/back env vars.

### Crisp

- Ensure website ID matches frontend env var.

### Grafana Cloud

- Confirm OTLP endpoint + Basic auth header.
- Validate traces arrive after a few API requests.

## 4) Smoke test (production)

1. Frontend loads and can reach backend:
   - Landing page renders platform status.
2. Backend readiness:
   - `GET /readyz` returns `{"status":"ready"}`
3. Auth:
   - Sign in via Clerk, then open `/app` and confirm `GET /api/v1/auth/me` works.
4. Billing:
   - Use `/pricing` and confirm checkout redirects, then confirm webhook updates subscription state.
5. Integrations:
   - PostHog events appear (frontend and backend).
   - Sentry captures test error (optional).
   - Crisp widget loads (optional).
   - Grafana Cloud receives traces (optional).
6. Background jobs:
   - Trigger a new-user sign-in to enqueue a welcome email job.
   - Confirm worker processes the job (Resend in prod, console locally).
