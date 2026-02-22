# Multi-tenant Model

This template uses an organization/workspace model as the default multi-tenant strategy.

## Tenant model

- A user can belong to one or more organizations.
- Product data belongs to an organization unless it is explicitly global platform metadata.
- Effective authorization is determined by membership and role within the active organization.

### Personal workspace (Option A)

Every new user gets a **personal workspace** implemented as a normal `organizations` row with:

- `kind = 'personal'`
- `personal_owner_user_id = <user_id>`

This personal workspace is enforced to be **single-member** (owner only). Users can still create and join other (team) organizations.

### Team organizations

Team organizations (`kind = 'team'`) support multiple members and roles.

- The `organization_members.role` column is constrained to: `owner`, `admin`, `member`.
- Team organizations are enforced to always have **at least one** `owner` (the last owner cannot be removed or demoted).

## Recommended tables

- `organizations`
  - `id`
  - `name`
  - `slug`
  - `created_at`, `updated_at`
- `organization_members`
  - `id`
  - `organization_id`
  - `user_id`
  - `role` (for example `owner`, `admin`, `member`)
  - `created_at`, `updated_at`

Every tenant-scoped business table must include `organization_id`.

## Authorization rules

- Resolve active organization context on each request.
- Verify membership before all tenant-scoped reads and writes.
- Apply role checks for sensitive actions (billing changes, member management, audit access, settings).
- Deny by default when organization context is missing or invalid.

### RBAC roles

Membership includes a `role`:

- `owner`: full control of an organization.
- `admin`: manage billing/settings and operational data.
- `member`: default role for day-to-day usage.

Role hierarchy: `owner` > `admin` > `member`.

Current API enforcement:

- Billing endpoints require `admin` or higher: `POST /api/v1/billing/checkout-session`, `POST /api/v1/billing/portal-session`.
- Audit events require `admin` or higher: `GET /api/v1/audit/events`.

## API scoping conventions

- Never accept raw `organization_id` from clients without server-side membership validation.
- Where possible, scope routes by organization context and infer tenant internally.
- Keep membership checks close to boundary handlers/middleware and re-check in service layer for sensitive operations.

## Query safety patterns

- Include `organization_id` in all tenant-scoped `WHERE` clauses.
- Avoid broad list queries that do not include tenant scoping.
- For joins, enforce tenant equality at query boundaries.

## Auditing requirements

Record audit events for:

- Organization creation/deletion
- Membership invites/acceptance/removal
- Role changes
- Cross-organization context switches (if supported in UI)

Audit events should include actor `user_id`, `organization_id`, action type, and timestamp.

## Common anti-patterns

- Using user-only filters for tenant-scoped data.
- Relying on frontend-selected organization without backend validation.
- Allowing wildcard admin endpoints to bypass tenant constraints unintentionally.
