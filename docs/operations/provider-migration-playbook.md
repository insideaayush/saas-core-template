# Provider Migration Playbook

This playbook describes how to migrate authentication providers with minimal user disruption.

## Objectives

- Preserve internal `user_id` and organization relationships.
- Avoid lock-in to provider-specific identifiers in domain tables.
- Minimize forced user actions during migration.

## Preconditions

- App uses an identity mapping model (`auth_identities`) separate from domain user records.
- Business logic depends on provider-agnostic auth interfaces.
- Email verification and identity-linking rules are enforced.

## Migration phases

## Phase 1: Preparation

- Add new provider adapter behind existing `AuthProvider` interface.
- Validate callback/session verification behavior in staging.
- Add migration telemetry dashboards (success/failure counters).
- Decide fallback policy for unresolved identities.

## Phase 2: Dual-run

- Keep legacy provider login enabled.
- Enable new provider login in parallel.
- On successful legacy-provider authentication:
  - Resolve internal `user_id`.
  - Create or update mapping for the new provider.
  - Mark identity as migrated.

This is just-in-time (JIT) migration and avoids a single high-risk bulk cutover.

## Phase 3: Catch-up

- Identify users not migrated via dual-run activity.
- Send targeted communication for sign-in refresh (magic link or social login preferred).
- If passwordless methods are unavailable, require controlled reset flow on new provider.

## Phase 4: Cutover

- Disable new sessions on old provider.
- Keep old mapping data read-only for audit/history.
- Monitor failed sign-ins and account-link events closely.

## Phase 5: Decommission

- Remove old provider runtime dependencies after stability period.
- Archive migration logs and incident records.
- Update architecture docs and runbook ownership.

## Account linking policy

- Never auto-link identities only by unverified email.
- Require provider-verified identity evidence for linking.
- For ambiguous cases, use support-assisted verification workflow.

## Password migration reality

- Many managed providers do not allow password hash export.
- Expect some users to set new passwords if you used password auth before migration.
- To minimize this risk in starter templates, prefer magic links/social/passkeys.

## Rollback plan

- Keep old provider sign-in path available during early cutover window.
- Toggle migration mode with feature flags.
- Maintain reversible adapter routing until key KPIs stabilize.

## KPIs to track

- Login success rate by provider
- Identity-link success/failure counts
- Support tickets related to sign-in/access
- Time-to-resolution for migration incidents
