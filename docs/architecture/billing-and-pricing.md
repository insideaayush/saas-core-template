# Billing and Pricing Model

This template treats the billing provider as a payment processor and event source, while application state controls product access.

## Ownership boundaries

- Provider-owned:
  - Payment methods, card tokens, processor transaction internals.
- App-owned:
  - Plans, plan entitlements, subscription status used for feature gating, organization billing linkage.

## Recommended tables

- `plans`
  - `id`
  - `code` (for example `free`, `pro`, `team`)
  - `display_name`
  - `billing_interval` (monthly, yearly)
  - `is_active`
- `subscriptions`
  - `id`
  - `organization_id`
  - `plan_id`
  - `status` (trialing, active, past_due, canceled, etc.)
  - `provider`
  - `provider_customer_id`
  - `provider_subscription_id`
  - `current_period_end`
  - `created_at`, `updated_at`

Use uniqueness constraints for provider mapping fields where appropriate.

## Pricing flow

1. User selects plan on pricing page.
2. Backend creates provider checkout session for the organization.
3. Provider redirects user through checkout.
4. Webhook event confirms completion.
5. App updates `subscriptions` and entitlements.
6. Product access checks rely on internal `subscriptions` state.

## Webhook requirements

- Verify webhook signatures.
- Implement idempotency (event ID tracking or equivalent safe upsert patterns).
- Support replay safely.
- Ignore unknown events safely while logging for review.

Minimum event coverage:

- Checkout completion
- Subscription created/updated/canceled
- Invoice payment success/failure (if used for entitlement transitions)

## Feature gating rules

- Never gate product features directly from client-side billing provider state.
- Evaluate entitlements from internal subscription records.
- Apply organization scope when evaluating paid features.

## Billing portal

- Allow organizations to self-manage billing via provider customer portal.
- Treat provider portal updates as asynchronous and reflect them through webhook sync.

## Anti-patterns to avoid

- Driving entitlements from provider API calls on every request.
- Treating webhook order as guaranteed.
- Using a single global subscription for all organizations in a multi-tenant app.
