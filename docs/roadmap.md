# Roadmap

## Phase 0: Shell (current)
- Deployable monorepo with Next.js frontend and Go backend.
- Postgres on Render and Upstash Redis integration.
- Local development parity using Docker Compose for Postgres and Redis.
- CI checks on `main`.

## Phase 0.5: Governance foundation
- `AGENTS.md` defines non-negotiable architecture, security, and migration guardrails.
- `docs/` playbooks define auth identity boundaries, tenant model, billing model, SOC 2 foundations, and provider migration workflows.
- All upcoming phases must follow provider-agnostic adapter boundaries and internal domain ownership rules.

## Phase 1: Thinking Wedge
- Notes capture with markdown and lightweight tagging.
- AI ideation workspace backed by conversation memory.
- Search and retrieval over notes.
- Must enforce tenant-aware access controls for any shared/multi-account features.

## Phase 2: Personal Ops
- Daily dashboard with routines and execution tracking.
- Health ingestion (manual plus selective integrations).
- Finance ingestion (manual plus CSV import).
- Must follow sensitive-data logging redaction and audit-event requirements from governance docs.

## Phase 3: Productization
- Multi-account auth model.
- Billing and account boundaries.
- Data portability and privacy controls.
- Implementation must use provider adapter interfaces and preserve internal `user_id` as the primary domain key.
