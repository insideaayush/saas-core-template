# Docs Index

This directory contains implementation playbooks for contributors and AI agents.

## First 30 minutes

1. Read `../AGENTS.md` for non-negotiable architecture and security guardrails.
2. Review the architecture playbooks in `docs/architecture/`.
3. Review the operations playbooks in `docs/operations/`.
4. Confirm your change keeps provider boundaries portable and tenant isolation strict.

## Architecture playbooks

- [Auth and Identity](architecture/auth-and-identity.md)
  - Internal user model, identity linking, provider boundary pattern.
- [Multi-tenant Model](architecture/multi-tenant-model.md)
  - Organization/workspace model and authorization checks.
- [Billing and Pricing](architecture/billing-and-pricing.md)
  - Internal subscription state, checkout, webhook synchronization.

## Operations playbooks

- [SOC 2 Foundations](operations/compliance-soc2-foundations.md)
  - Baseline controls and evidence expectations.
- [Provider Migration Playbook](operations/provider-migration-playbook.md)
  - Dual-run, just-in-time migration, and cutover strategy.

## Contribution checklist

Before opening a PR:

- Confirm all new data writes are tenant scoped.
- Confirm provider SDK usage is isolated behind interfaces/adapters.
- Confirm logs do not include secrets or sensitive payloads.
- Add/update docs for any auth, tenancy, billing, or compliance-sensitive change.
