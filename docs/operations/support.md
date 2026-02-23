# Support (Crisp)

This template supports an optional support widget in the frontend.

## Local development

Default is disabled:

- `NEXT_PUBLIC_SUPPORT_PROVIDER=none`

## Production (Crisp)

Set:

- `NEXT_PUBLIC_SUPPORT_PROVIDER=crisp`
- `NEXT_PUBLIC_CRISP_WEBSITE_ID=<website id>`

## Switching providers

Support widgets should be loaded and controlled through an internal client boundary so swaps are localized to the adapter.
