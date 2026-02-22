# Audit Logs

This template includes an `audit_events` table for security-relevant and business-relevant actions.

## Local development

Audit events are written to Postgres.

## API

- `GET /api/v1/audit/events` (org-scoped) returns recent audit events for the active organization.

## What should be audited

At minimum:

- Identity and access changes (sign-in, role changes, invites)
- Billing actions (checkout/portal sessions, subscription changes)
- File uploads and sensitive operations

