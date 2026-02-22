# AGENTS.md

This file defines non-negotiable engineering guardrails for humans and AI agents working in this repository.

## Product intent

- This repository is a startup starter template.
- It must be quick to launch while preserving portability across auth and billing providers.
- Domain ownership stays in application storage even when managed providers are used.

## Template initialization protocol

- Before feature work on a new clone, initialize project naming:
  - Run `./scripts/init-template.sh "<new-project-name>"`.
  - Use a lowercase kebab-case name (for example `acme-core`).
- The script must be run from the repository root.
- After initialization, verify rename integrity:
  - Run `rg "saas-core-template|saas_core_template"`.
  - Expected remaining matches should be only intentional references (for example instructions in `scripts/init-template.sh` or docs explaining initialization).
- If unintended template-name references remain, replace them before implementing product changes.
- Do not manually edit Go module/import paths first; always run the init script before any manual rename fixes.

## Agent operating workflow

- Follow this order for all implementation tasks:
  1. Read `AGENTS.md` and relevant playbooks in `docs/`.
  2. Load relevant project skills from `.cursor/skills/` for design/implementation/review tasks.
  3. Confirm naming initialization is complete.
  4. Inspect existing code before designing changes.
  5. Implement minimal, scope-limited changes.
  6. Run validation commands.
  7. Update docs when behavior/contracts changed.
- Prefer small, composable edits over broad rewrites.
- Preserve existing architecture boundaries and avoid introducing hidden coupling.
- If requirements conflict with these guardrails, stop and ask for explicit product/architecture direction.

## Required pre-flight checks

- Verify repository naming:
  - `rg "saas-core-template|saas_core_template"` and confirm only intentional matches.
- Confirm environment setup files remain aligned:
  - `.env.example`
  - `backend/.env.example`
  - `frontend/.env.example`
- Confirm changes are scoped to the requested task and do not include unrelated refactors.

## Validation gates before completion

- Backend:
  - Run `go test ./...` in `backend/` when possible.
  - Run `gofmt` on changed Go files.
- Frontend:
  - Run `npm run lint` and `npm run typecheck` in `frontend/`.
  - Run `npm run build` for route/config changes.
- Cross-cutting:
  - Re-run targeted searches for old identifiers and stale provider references.
  - Verify no secrets were added to tracked files.
  - Confirm managed integrations remain optional and local E2E works with console/noop defaults (telemetry, analytics, error reporting, support).

## Git and change hygiene

- Never commit `.env` or secret material.
- Keep commit messages focused on intent and impact.
- Do not bundle unrelated concerns in the same commit.
- Do not force-push protected branches.
- If push/auth fails locally, report exact failure and next remediation commands.
- Apply `.cursor/skills/saas-git-workflow` for branch selection, PR routing, and release/version decisions.
- Follow branch policy:
  - `main` is release-only.
  - `develop` is integration branch.
  - `dev` or `feature/*` is for feature implementation.
- Keep `VERSION` as the template semantic version source of truth.

## Documentation update policy

- Any change to auth, tenancy, billing, migrations, deployment, or initialization must update docs in the same task.
- Any change to business/domain behavior must update business docs according to `docs/architecture/business-logic-documentation.md`.
- Required docs touchpoints by change type:
  - Auth identity flow: `docs/architecture/auth-and-identity.md`
  - Tenant model/rules: `docs/architecture/multi-tenant-model.md`
  - Billing flow/webhooks: `docs/architecture/billing-and-pricing.md`
  - Domain/business rules organization and lifecycle: `docs/architecture/business-logic-documentation.md`
  - Control/evidence posture: `docs/operations/compliance-soc2-foundations.md`
  - Provider swap/cutover behavior: `docs/operations/provider-migration-playbook.md`
  - Agent protocol changes: `AGENTS.md` and `docs/operations/agent-workflow.md`

## Default implementation preferences

- Keep provider-specific logic inside adapters/services; do not leak SDK-specific shapes into domain models.
- Prefer explicit DTOs and typed request/response contracts at API boundaries.
- Keep tenant authorization checks centralized and deny-by-default.
- Make webhook/event handlers idempotent and replay-safe.
- For Go backend changes, apply `.cursor/skills/go-saas-patterns`.
- For TypeScript/Next.js changes, apply `.cursor/skills/typescript-saas-patterns`.

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
