# AGENTS.md

This file defines non-negotiable engineering guardrails for humans and AI agents working in this repository.

## Product intent

- This repository is a startup starter template.
- It must be quick to launch while preserving portability across auth and billing providers.
- Domain ownership stays in application storage even when managed providers are used.

## Hard architecture boundaries

- Use internal `user_id` as the primary identity key across all domain tables.
- Store provider identifiers only in mapping fields/tables (for example `provider_user_id`, `provider_org_id`, `provider_customer_id`).
- Never use provider IDs as foreign keys in core business tables.
- Design adapter interfaces so provider swaps do not require business logic rewrites.

## Data ownership model

- Provider-owned data:
  - Credentials, password hashes, MFA factors, auth sessions, OAuth internals.
  - Payment method details, card tokens, payment processor internals.
- App-owned data:
  - Users, organizations, membership/roles, feature access, plans, subscription state, usage counters.
  - Audit events, tenant boundaries, product configuration.

## Authentication rules

- Authentication provider verifies identity; app authorizes access.
- Email must be treated as an attribute, not a stable identity key.
- Support multiple identities per user through an identity-link model.
- Require verified identity before linking a new auth method to an existing user.

## Multi-tenant rules

- All application data access must be tenant scoped.
- Every tenant-scoped table must include a tenant key (for example `organization_id`).
- Every read/write path must enforce membership checks for the active tenant.
- Cross-tenant access is forbidden unless explicitly required for platform admin features.

## Billing rules

- Billing provider events are inputs; internal subscription state is the source of truth for feature gating.
- Webhook handlers must be idempotent and safe to replay.
- Keep a clear mapping from internal account/org to billing customer/subscription IDs.

## API and service conventions

- Business services depend on interfaces, not concrete provider SDKs.
- Keep provider-specific code isolated in adapter packages.
- Route handlers must avoid direct provider SDK calls except via service/adapters.
- Use explicit error types for authorization, validation, and dependency failures.

## Logging, audit, and sensitive data

- Emit structured logs for key auth, tenant, and billing events.
- Redact secrets and sensitive fields in logs by default.
- Capture audit events for security-relevant actions:
  - Sign in/out
  - Identity linking/unlinking
  - Membership and role changes
  - Plan/subscription changes

## SOC 2 foundations (default)

- Enforce least privilege and role-based access patterns.
- Keep production configuration in environment variables and secret stores only.
- Require CI checks and code review before merge to protected branches.
- Track security-sensitive changes and operational incidents.

## Migration safety requirements

- Implement an identity mapping model that supports dual-provider periods.
- Avoid assuming password export/import support between providers.
- Ensure account linking and migration can be done using verified identifiers.
- Keep migration playbooks updated in `docs/operations/`.

## Documentation contract

- Any change to auth, tenancy, billing, or compliance-sensitive behavior must update relevant docs in `docs/`.
- If a change introduces a new provider dependency, document boundaries and migration implications.
