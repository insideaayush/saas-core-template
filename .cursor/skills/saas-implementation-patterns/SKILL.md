---
name: saas-implementation-patterns
description: Applies proven SaaS implementation patterns for auth, multi-tenancy, billing, webhooks, and migration-safe boundaries. Use when implementing backend or frontend features in SaaS apps, especially account models, authorization, subscription flows, and provider integrations.
---

# SaaS Implementation Patterns

Use this skill when writing code so SaaS features remain secure, tenant-safe, and migration-friendly.

## Core patterns

## 1) Identity mapping pattern

- Keep `users` as internal identity table.
- Link providers through mapping table (for example `auth_identities` with `provider`, `provider_user_id`).
- Treat email as attribute, not immutable identity key.

## 2) Tenant authorization pattern

- Resolve active tenant context at request boundary.
- Enforce membership before all tenant-scoped operations.
- Add tenant key to tenant-owned tables and query filters.

## 3) Billing ownership pattern

- Provider checkout/portal handles payments.
- Internal `subscriptions` state controls feature access.
- Webhooks update internal state asynchronously.

## 4) Idempotent webhook pattern

- Verify signatures.
- Parse event type safely.
- Upsert by provider-stable IDs.
- Allow replay without duplicate side effects.

## 5) Provider abstraction pattern

- Define provider interfaces (`AuthProvider`, `BillingProvider`).
- Keep SDK-specific code in adapters.
- Domain services consume interfaces only.

## Implementation checklist

Copy this checklist into your task and mark as done:

```text
- [ ] Domain data keyed by internal IDs
- [ ] Provider IDs stored in mapping fields/tables only
- [ ] Tenant scoping enforced in handlers and queries
- [ ] Webhook handlers idempotent and signature-verified
- [ ] Sensitive values redacted from logs
- [ ] Docs updated for behavior/contract changes
```

## Output expectations for code tasks

- Explain where auth boundary lives.
- Explain where tenant enforcement occurs.
- Explain how billing state is synchronized.
- List migration implications if provider-coupled code was introduced.

## Default decisions for this repository

- Managed-first auth/billing for delivery speed.
- SOC 2 foundation controls from day one.
- Migration-safe structure over short-term convenience hacks.

## Reject these changes

- Direct provider SDK usage spread across route handlers.
- Entitlement gating based solely on provider API calls.
- Mixed tenant data queries without `organization_id` filter.
