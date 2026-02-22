# Error Reporting (Sentry)

This template supports error reporting with local console defaults and managed provider opt-in.

## Local development

- Backend defaults to `ERROR_REPORTING_PROVIDER=console` and logs captured exceptions.
- Frontend defaults to `NEXT_PUBLIC_ERROR_REPORTING_PROVIDER=console` and logs captured exceptions in the browser console.

## Production (Sentry)

### Backend (Go)

Set:

- `ERROR_REPORTING_PROVIDER=sentry`
- `SENTRY_DSN=<dsn>`
- `SENTRY_ENVIRONMENT=production`

### Frontend (Next.js)

Set:

- `NEXT_PUBLIC_ERROR_REPORTING_PROVIDER=sentry`
- `NEXT_PUBLIC_SENTRY_DSN=<dsn>`
- `NEXT_PUBLIC_SENTRY_ENVIRONMENT=production`
- `NEXT_PUBLIC_APP_VERSION=<release identifier>` (optional)

## Switching providers

Keep provider-specific SDK calls behind an internal adapter boundary. Only env vars should change when swapping providers.
