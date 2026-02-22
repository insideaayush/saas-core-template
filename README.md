# saas-core-template

Production-shaped foundation for launching a startup SaaS baseline quickly.

## Stack
- Frontend: Next.js (TypeScript)
- UI: shadcn/ui (Tailwind + Radix)
- Backend: Go (`net/http`)
- Database: Postgres
- Cache: Redis (Upstash in cloud, local Redis in development)
- Auth: Clerk (managed auth, organization context)
- Billing: Stripe (checkout + portal + webhook sync)
- Deploy: Render (backend + Postgres) + Vercel (frontend)
- CI: GitHub Actions

## Repository layout
- `frontend/` Next.js shell and app routes
- `backend/` Go API shell, auth/billing endpoints, and migrations
- `docker-compose.yml` local Postgres + Redis
- `render.yaml` Render blueprint for backend and Postgres
- `.github/workflows/ci.yml` CI for frontend and backend
- `docs/roadmap.md` product phases after shell

## Prerequisites
- Node.js 20+
- npm 10+
- Go 1.22+
- Docker Desktop (or Docker Engine + Compose)

## Environment variables
Copy examples and adjust as needed:

```bash
cp .env.example .env
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
```

Core variables:
- Backend
  - `PORT` (default `8080`)
  - `DATABASE_URL` (local Postgres or Render Postgres URL)
  - `REDIS_URL` (local Redis or Upstash Redis URL)
  - `APP_BASE_URL` (frontend URL used for checkout return paths)
  - `APP_ENV` (`development` or `production`)
  - `APP_VERSION` (`dev`, commit SHA, or release tag)
  - `OTEL_SERVICE_NAME` (default `saas-core-template-backend`)
  - `OTEL_TRACES_EXPORTER` (`console`, `otlp`, or `none`)
  - `OTEL_EXPORTER_OTLP_ENDPOINT` (local collector default `http://localhost:4318`)
  - `OTEL_EXPORTER_OTLP_HEADERS` (for managed OTLP auth, e.g. Grafana Cloud)
  - `ERROR_REPORTING_PROVIDER` (`console`, `sentry`, or `none`)
  - `SENTRY_DSN` (backend error reporting)
  - `SENTRY_ENVIRONMENT` (defaults to empty)
  - `ANALYTICS_PROVIDER` (`console`, `posthog`, or `none`)
  - `POSTHOG_PROJECT_KEY`
  - `POSTHOG_HOST`
  - `CLERK_SECRET_KEY`
  - `CLERK_API_URL` (default `https://api.clerk.com`)
  - `STRIPE_SECRET_KEY`
  - `STRIPE_WEBHOOK_SECRET`
  - `STRIPE_API_URL` (default `https://api.stripe.com/v1`)
  - `STRIPE_PRICE_PRO_MONTHLY`
  - `STRIPE_PRICE_TEAM_MONTHLY`
- Frontend
  - `NEXT_PUBLIC_API_URL` (e.g. `http://localhost:8080`)
  - `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`
  - `NEXT_PUBLIC_ANALYTICS_PROVIDER` (`console`, `posthog`, or `none`)
  - `NEXT_PUBLIC_POSTHOG_KEY`
  - `NEXT_PUBLIC_POSTHOG_HOST`
  - `NEXT_PUBLIC_SUPPORT_PROVIDER` (`crisp` or `none`)
  - `NEXT_PUBLIC_CRISP_WEBSITE_ID`
  - `NEXT_PUBLIC_ERROR_REPORTING_PROVIDER` (`console`, `sentry`, or `none`)
  - `NEXT_PUBLIC_SENTRY_DSN`
  - `NEXT_PUBLIC_SENTRY_ENVIRONMENT`
  - Locale is stored in a `locale` cookie (supported: `en`, `es`)

## Database migrations

SQL migrations live in `backend/migrations/`.

Apply them with your preferred migration tool before using auth/billing endpoints.
Initial migration file:

- `backend/migrations/0001_identity_tenancy_billing.up.sql`

## Local development
Run infra first:

```bash
make infra-up
```

This starts Postgres, Redis, and a local OpenTelemetry collector (for local tracing).

Start backend in one terminal:

```bash
make dev-api
```

Start frontend in another terminal:

```bash
make dev-ui
```

Open:
- Frontend: `http://localhost:3000`
- Backend health: `http://localhost:8080/healthz`
- Backend readiness: `http://localhost:8080/readyz`
- Backend metadata: `http://localhost:8080/api/v1/meta`
- Sign in: `http://localhost:3000/sign-in`
- Pricing: `http://localhost:3000/pricing`

Stop local infra:

```bash
make infra-down
```

## CI checks
Run locally:

```bash
make ci
```

GitHub Actions workflow runs on pull requests and pushes to `main`, `develop`, and `dev`:
- Backend: `go test`, `go vet`, build
- Frontend: install, lint, typecheck, build

CI currently runs on `main`, `develop`, and `dev` branches.

## Branch strategy

- `main`: release branch
- `develop`: integration branch
- `dev` or `feature/*`: feature implementation branches

Recommended flow:

1. Build in `dev` or `feature/*`.
2. Open PR into `develop`.
3. Release from `develop` into `main`.

## Versioning

- Template version source of truth: `VERSION`
- Versioning scheme: SemVer (`MAJOR.MINOR.PATCH`)
- CI validates `VERSION` format on `main`, `develop`, and `dev`.

See `docs/operations/git-branching-and-versioning.md` for full guidance.

## Production deployment

This template deploys:
- Backend + Postgres on Render (see `render.yaml`)
- Frontend on Vercel (deploy `frontend/`)

### Backend + Postgres (Render)
1. Connect this GitHub repository in Render.
2. Create services from the `render.yaml` blueprint.
3. Set backend secrets: `REDIS_URL`, `CLERK_SECRET_KEY`, `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, Stripe price IDs.
4. Set `APP_BASE_URL` to your Vercel frontend URL (used for Stripe return URLs).
5. Ensure auto-deploy is enabled for `main`.

### Frontend (Vercel)
1. Import the repo in Vercel.
2. Set project root directory to `frontend/`.
3. Set environment variables:
   - `NEXT_PUBLIC_API_URL` = Render backend URL
   - `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` = Clerk publishable key
4. Deploy.

Deployment flow:
- Push to `main` -> GitHub Actions CI passes -> Render auto-deploys backend; Vercel deploys frontend.

## Notes
- Local Redis exists for parity; production uses Upstash Redis.
- `main` branch protection should require CI checks before merge.
- Auth/billing provider boundaries and migration playbooks are documented in `docs/`.
## Initialize from template

After cloning, run:

```bash
./scripts/init-template.sh "your-project-name"
```

This replaces `saas-core-template` references across tracked files (including Go module/import paths and deployment service names) and prints follow-up commands.
