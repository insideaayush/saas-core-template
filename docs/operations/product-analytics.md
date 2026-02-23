# Analytics (PostHog)

This template supports analytics via provider boundaries in both the frontend and backend.

## Local development

Defaults are safe for local end-to-end runs:

- `NEXT_PUBLIC_ANALYTICS_PROVIDER=console` logs analytics calls in the browser console.
- Set `NEXT_PUBLIC_ANALYTICS_PROVIDER=none` to disable.
- Backend defaults to `ANALYTICS_PROVIDER=console` and logs events via structured logs.

## Production (PostHog Cloud)

Configure the frontend env vars:

- `NEXT_PUBLIC_ANALYTICS_PROVIDER=posthog`
- `NEXT_PUBLIC_POSTHOG_KEY=<project api key>`
- `NEXT_PUBLIC_POSTHOG_HOST=https://app.posthog.com` (or your PostHog instance URL)

Configure the backend env vars:

- `ANALYTICS_PROVIDER=posthog`
- `POSTHOG_PROJECT_KEY=<project api key>`
- `POSTHOG_HOST=https://app.posthog.com` (or your PostHog instance URL)

## Switching providers

Analytics calls should use the internal client boundary (not provider SDKs directly). This keeps provider swaps localized to the integration adapter.
