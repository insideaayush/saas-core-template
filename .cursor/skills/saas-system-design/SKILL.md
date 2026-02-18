---
name: saas-system-design
description: Designs SaaS architecture from requirements with clear boundaries, tradeoffs, and rollout plans. Use when the user asks for system design, architecture decisions, scalability planning, tenant isolation strategy, auth/billing architecture, or technical design docs.
---

# SaaS System Design

Use this skill to convert product goals into implementation-ready SaaS architecture decisions.

## Inputs to gather first

- Product scope and primary workflows.
- Tenant model (`single-tenant`, `org/workspace`, `hybrid`).
- Compliance posture (SOC 2 baseline, HIPAA-aware, other constraints).
- Launch priority (`speed` vs `control` vs `cost`).
- Current stack constraints (language/framework/cloud/provider lock-in tolerance).

If key inputs are missing, ask 1-2 critical clarifying questions before proposing architecture.

## Design workflow

1. Define domain boundaries
   - Separate identity/authentication, authorization, tenant ownership, billing, and product domain.
2. Define data ownership
   - Specify provider-owned vs app-owned data explicitly.
3. Define control-plane vs data-plane
   - Control-plane: account setup, billing, admin settings.
   - Data-plane: tenant-scoped product operations.
4. Choose integration model
   - Managed-first with adapter boundaries by default unless user asks for self-hosted control.
5. Define reliability and security baselines
   - Idempotency, replay safety, audit events, sensitive logging redaction.
6. Define phased delivery
   - MVP path, scale path, migration path.

## Required output format

Use this structure:

```markdown
## Context
- <constraints and goals>

## Architecture Decisions
- <decision>: <why>

## Boundaries
- Auth:
- Tenancy:
- Billing:
- Data ownership:

## Tradeoffs
- Option A / Option B

## Phased rollout
1. Phase 1
2. Phase 2

## Risks and mitigations
- <risk>: <mitigation>
```

Add a mermaid diagram when service boundaries or data flows are non-trivial.

## SaaS defaults for this repository

- Internal `user_id` is the primary identity key.
- Tenant-scoped data must include tenant key and membership checks.
- Billing events are inputs; internal subscription state drives entitlements.
- Provider SDK usage should be isolated to adapters/services.

## Anti-patterns to reject

- Provider IDs as primary foreign keys in core domain tables.
- Authorization based only on provider metadata without internal membership checks.
- Non-idempotent webhook handlers.
- Cross-tenant reads/writes without explicit platform-admin boundary.
