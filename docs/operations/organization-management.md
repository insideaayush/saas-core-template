# Organization Management

This template uses app-owned organizations (workspaces) and membership roles stored in Postgres.

Clerk is used for authentication only; organization context and RBAC are enforced by the API using app-owned tables.

## Concepts

- Personal workspace: created automatically on first sign-in (`kind = 'personal'`), enforced single-member owner-only.
- Team organization: created by a signed-in user (`kind = 'team'`), supports multiple members and roles.
- Role hierarchy: `owner` > `admin` > `member`.

## API endpoints

All endpoints require a Clerk bearer token (`Authorization: Bearer ...`).

Organization context is selected via `X-Organization-ID: <internal org uuid>` for org-scoped endpoints.

- `GET /api/v1/orgs`: list organizations the user belongs to (includes role + kind).
- `POST /api/v1/orgs`: create a new team organization.
- `GET /api/v1/org/members`: list members for the active org (admin+).
- `POST /api/v1/org/invites`: create an invite for the active org (admin+, team orgs only).
- `POST /api/v1/org/invites/accept`: accept an invite token (email must match the signed-in user).
- `PATCH /api/v1/org/members/{userId}`: change a member role (owner-only, team orgs only).
- `DELETE /api/v1/org/members/{userId}`: remove a member (owner-only, team orgs only).

## Invite flow

1. Owner/admin creates an invite for a team org via `POST /api/v1/org/invites`.
2. The API returns an `acceptUrl` pointing at `GET /app/invite?token=...` and (if the worker is enabled) enqueues an email job to deliver the link.
3. The invited user signs in, opens the link, and the UI calls `POST /api/v1/org/invites/accept`.

## Active organization selection (frontend)

The frontend stores the active org UUID in `localStorage` under `activeOrganizationId` and sends it as `X-Organization-ID`.

