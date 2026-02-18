# Novame WebOS Shell

Production-shaped foundation for a personal WebOS project.

## Stack
- Frontend: Next.js (TypeScript)
- Backend: Go (`net/http`)
- Database: Postgres
- Cache: Redis (Upstash in cloud, local Redis in development)
- Auth: Clerk (managed auth, organization context)
- Billing: Stripe (checkout + portal + webhook sync)
- Deploy: Render
- CI: GitHub Actions

## Repository layout
- `frontend/` Next.js shell and app routes
- `backend/` Go API shell, auth/billing endpoints, and migrations
- `docker-compose.yml` local Postgres + Redis
- `render.yaml` Render blueprint for backend, frontend, and Postgres
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

GitHub Actions workflow runs on pull requests and pushes to `main`:
- Backend: `go test`, `go vet`, build
- Frontend: install, lint, typecheck, build

## Render deployment
`render.yaml` defines:
- `novame-backend` (Go web service)
- `novame-frontend` (Node web service for Next.js)
- `novame-postgres` (managed Postgres)

### Setup steps on Render
1. Connect this GitHub repository in Render.
2. Create services from `render.yaml` blueprint.
3. Set backend secrets: `REDIS_URL`, `CLERK_SECRET_KEY`, `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, Stripe price IDs.
4. Set frontend `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`.
5. Confirm frontend `NEXT_PUBLIC_API_URL` points to backend service URL.
6. Ensure auto-deploy is enabled for `main`.

Deployment flow:
- Push to `main` -> GitHub Actions CI passes -> Render auto-deploys updated services.

## Notes
- Local Redis exists for parity; production uses Upstash Redis.
- `main` branch protection should require CI checks before merge.
- Auth/billing provider boundaries and migration playbooks are documented in `docs/`.
# saas-core-template
