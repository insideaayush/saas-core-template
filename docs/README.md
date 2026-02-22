# Docs Index

This directory contains implementation playbooks for contributors and AI agents.

## Start here

- `overview.md` gives a reading order and repo map.

## First 30 minutes

1. Run `./scripts/init-template.sh "<new-project-name>"` if this is a new clone from template.
2. Read `../AGENTS.md` for non-negotiable architecture and security guardrails.
3. Review the architecture playbooks in `docs/architecture/`.
4. Review the operations playbooks in `docs/operations/`.
5. Confirm your change keeps provider boundaries portable and tenant isolation strict.
6. Skim `../README.md` for local + production deployment setup (Render backend + Postgres, Vercel frontend).

## Architecture playbooks

- [Auth and Identity](architecture/auth-and-identity.md)
  - Internal user model, identity linking, provider boundary pattern.
- [Multi-tenant Model](architecture/multi-tenant-model.md)
  - Organization/workspace model and authorization checks.
- [Billing and Pricing](architecture/billing-and-pricing.md)
  - Internal subscription state, checkout, webhook synchronization.
- [Business Logic Documentation](architecture/business-logic-documentation.md)
  - Canonical structure and maintenance rules for domain/business docs.

## Operations playbooks

- [SOC 2 Foundations](operations/compliance-soc2-foundations.md)
  - Baseline controls and evidence expectations.
- [Production Setup Checklist](operations/production-setup-checklist.md)
  - End-to-end deployment wiring (Render + Vercel + providers).
- [Observability (OpenTelemetry)](operations/observability.md)
  - Local tracing collector and production export configuration.
- [Product Analytics (PostHog)](operations/product-analytics.md)
  - Local console analytics and managed PostHog configuration.
- [Error Reporting (Sentry)](operations/error-reporting.md)
  - Local console error capture and managed Sentry configuration.
- [Support (Crisp)](operations/support.md)
  - Optional support widget integration and provider swaps.
- [Provider Migration Playbook](operations/provider-migration-playbook.md)
  - Dual-run, just-in-time migration, and cutover strategy.
- [Agent Workflow Runbook](operations/agent-workflow.md)
  - Standard protocol for agent-assisted planning, implementation, validation, and delivery.
- [Git Branching and Versioning](operations/git-branching-and-versioning.md)
  - Branch strategy (`main`/`develop`/`dev`) and SemVer release flow.

## Project skills

Project-scoped Cursor skills live in `.cursor/skills/` and should be used by agents proactively:

- `saas-system-design`
  - Architecture decision workflows and SaaS boundary design.
- `saas-implementation-patterns`
  - Implementation patterns for auth, tenancy, billing, webhooks, and migration safety.
- `saas-architecture-review-gate`
  - Pre-merge architecture/risk review checklist and findings format.
- `go-saas-patterns`
  - Go backend patterns for readable, testable, and extensible service architecture.
- `typescript-saas-patterns`
  - TypeScript/Next.js patterns for typed boundaries, composable modules, and testable UI/domain logic.
- `saas-git-workflow`
  - Branch selection, PR routing, commit hygiene, and template versioning decisions.

## Contribution checklist

Before opening a PR:

- Confirm all new data writes are tenant scoped.
- Confirm provider SDK usage is isolated behind interfaces/adapters.
- Confirm logs do not include secrets or sensitive payloads.
- Add/update docs for any auth, tenancy, billing, or compliance-sensitive change.

## Frontend guides

- [UI Design Guide](frontend/ui-design-guide.md)
  - Tailwind + Radix component conventions and UI principles.
