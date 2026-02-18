# Auth and Identity Model

This template separates identity verification from application authorization.

## Core principle

- Provider authenticates the user.
- Application authorizes actions using internal domain records.

## Ownership boundaries

- Provider-owned:
  - Password hashes, MFA material, social login token internals, session internals.
- App-owned:
  - Internal users, profile metadata used by product logic, tenant membership, roles, audit trail.

## Recommended schema shape

- `users`
  - `id` (internal UUID, primary key)
  - `primary_email`
  - `display_name`
  - `created_at`, `updated_at`
- `auth_identities`
  - `id`
  - `user_id` (FK to `users.id`)
  - `provider` (for example `clerk`, `authjs`, `keycloak`)
  - `provider_user_id` (provider-stable user identifier)
  - `provider_email`
  - `email_verified_at`
  - `created_at`, `updated_at`

Use a uniqueness constraint on (`provider`, `provider_user_id`).

## Identity-linking rules

- Never use email as the only join key for account linkage.
- Only link identities after provider-verified identity checks.
- Preserve historical provider IDs for audit and migration support.

## Adapter boundary

Application services should depend on an auth interface, not concrete SDKs.

Example interface:

```go
type AuthProvider interface {
    VerifyRequest(ctx context.Context, authHeader string) (VerifiedPrincipal, error)
    GetUser(ctx context.Context, providerUserID string) (ProviderUser, error)
}
```

`VerifiedPrincipal` should include provider name and stable provider user ID. The app then maps that identity into `users` and `auth_identities`.

## Request flow

1. Request arrives with provider token/session artifact.
2. Auth adapter verifies artifact with provider.
3. App loads/creates `auth_identities` mapping row.
4. App resolves internal `user_id` and tenant context.
5. Authorization checks run against internal tables.

## Passwordless recommendation

For portability, prefer passwordless methods (magic links, social auth, passkeys) over local password UX. This minimizes migration pain if provider changes later.

## Anti-patterns to avoid

- Directly using provider user ID as primary key in domain tables.
- Embedding provider SDK calls in unrelated business services.
- Building authorization logic from provider metadata only.
