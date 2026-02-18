# Business Logic Documentation

This guide defines how project business logic documentation is organized and maintained.

## Goals

- Keep domain behavior explicit and discoverable.
- Make code-to-doc traceability easy for agents and humans.
- Prevent drift between implementation and product rules.

## Folder structure

Use this structure under `docs/business/`:

- `docs/business/README.md`
  - Index of all business domains and owner mapping.
- `docs/business/<domain>/overview.md`
  - Domain purpose, core entities, lifecycle, and invariants.
- `docs/business/<domain>/rules.md`
  - Decision rules, constraints, edge cases, and policy logic.
- `docs/business/<domain>/flows.md`
  - User and system flows (state transitions, event interactions).
- `docs/business/<domain>/glossary.md`
  - Shared terms to keep naming consistent.

Suggested initial domains:

- `identity-access`
- `organizations-tenancy`
- `billing-entitlements`
- `workflows-product-core`

## Required content per domain

Each domain must document:

- **Intent**: what the domain exists to achieve.
- **Source of truth**: tables/services that own state.
- **Invariants**: conditions that must always hold.
- **State model**: statuses and valid transitions.
- **Authorization model**: who can do what and where checks happen.
- **Failure handling**: retries, idempotency, and compensating behavior.

## Maintenance rules

- Update docs in the same PR when business rules or behavior change.
- Every rule in code should map to a domain doc section.
- Every new domain feature must include:
  - updated `overview.md`
  - updated `rules.md` or `flows.md`
  - changelog note in `docs/business/README.md`
- If behavior changed but docs were not updated, treat review as incomplete.

## Ownership and review

- Assign a primary owner per domain in `docs/business/README.md`.
- Review checklist for PRs touching domain logic:
  - docs updated
  - invariants still valid
  - edge cases documented
  - naming consistent with glossary

## Change log convention

At the top of each domain doc, include:

```markdown
Last updated: YYYY-MM-DD
Updated by: <team/owner>
Related PR: <link or id>
```

## Anti-patterns

- Keeping business rules only in code comments.
- Mixing transport/API details into domain rule docs.
- Scattering policy notes across unrelated docs without a canonical domain file.
