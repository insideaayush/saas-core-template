---
name: saas-architecture-review-gate
description: Reviews SaaS code changes against architecture guardrails before merge. Use when the user asks for review, PR readiness, regression checks, or validation of auth, tenancy, billing, compliance, and migration safety.
---

# SaaS Architecture Review Gate

Use this skill to perform a focused architecture and risk review before merge.

## Review priorities (in order)

1. Correctness and regression risk
2. Tenant isolation and authorization safety
3. Auth and identity boundary integrity
4. Billing and webhook correctness
5. Compliance/logging posture (SOC 2 baseline)
6. Migration safety and provider coupling

## Required review checklist

```text
- [ ] No cross-tenant access paths introduced
- [ ] Internal IDs remain source of truth for domain relationships
- [ ] Provider integrations remain behind adapters/interfaces
- [ ] Webhook handling is idempotent and replay-safe
- [ ] Sensitive data is not logged
- [ ] Env/config docs updated for new variables
- [ ] Docs/playbooks updated when behavior changed
```

## Findings format

Output findings first, ordered by severity:

- `Critical`: must fix before merge
- `High`: strong risk, should fix before merge
- `Medium`: correctness/maintainability concern
- `Low`: minor issue or future hardening

For each finding include:

- What is wrong
- Why it matters
- Where it occurs (file path/symbol)
- Suggested fix

If no findings, explicitly state that and list residual risks or testing gaps.

## Fast fail conditions

Immediately flag as critical if any of these appear:

- Tenant-scoped query without tenant filter.
- Provider ID used as primary foreign key in domain data.
- Missing webhook signature verification.
- Secrets/tokens logged or committed.

## Repository-specific guardrails

- `AGENTS.md` is authoritative for guardrails.
- `docs/architecture/*` and `docs/operations/*` define expected patterns.
- Any deviation must be documented and explicitly approved in review notes.
