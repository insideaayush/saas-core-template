# Email

This template supports sending transactional emails via a provider adapter.

## Local development

Defaults are local-first:

- `EMAIL_PROVIDER=console` logs emails instead of sending.

## Production (Resend)

Set backend env vars:

- `EMAIL_PROVIDER=resend`
- `EMAIL_FROM=<verified sender, e.g. "Acme <no-reply@acme.com>">`
- `RESEND_API_KEY=<api key>`

Email sending is executed via background jobs (see `background-jobs.md`).

## Provider boundaries

- Keep provider-specific API calls inside `backend/internal/email/*`.
- Use jobs to ensure email sending is retryable and non-blocking.

