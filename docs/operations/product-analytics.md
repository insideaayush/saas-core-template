# Product Analytics (PostHog)

This template supports product analytics via a provider boundary in the frontend.

## Local development

Defaults are safe for local end-to-end runs:

- `NEXT_PUBLIC_ANALYTICS_PROVIDER=console` logs analytics calls in the browser console.
- Set `NEXT_PUBLIC_ANALYTICS_PROVIDER=none` to disable.

## Production (PostHog Cloud)

Configure the frontend env vars:

- `NEXT_PUBLIC_ANALYTICS_PROVIDER=posthog`
- `NEXT_PUBLIC_POSTHOG_KEY=<project api key>`
- `NEXT_PUBLIC_POSTHOG_HOST=https://app.posthog.com` (or your PostHog instance URL)

## Switching providers

Analytics calls should use the internal client boundary (not provider SDKs directly). This keeps provider swaps localized to the integration adapter.
