# Observability (OpenTelemetry)

This template supports OpenTelemetry tracing in the Go backend.

## Local development

`make infra-up` starts a local OpenTelemetry Collector (`otel-collector`) in `docker-compose.yml` that accepts OTLP:

- OTLP HTTP: `http://localhost:4318`
- OTLP gRPC: `localhost:4317`

Local defaults:

- Backend uses `OTEL_TRACES_EXPORTER=console` to print spans to stdout.
- To send traces to the local collector, set `OTEL_TRACES_EXPORTER=otlp`.

## Production (Grafana Cloud)

To export traces directly to Grafana Cloud (no collector required), configure:

- `OTEL_TRACES_EXPORTER=otlp`
- `OTEL_EXPORTER_OTLP_ENDPOINT=<your Grafana Cloud OTLP endpoint>`
- `OTEL_EXPORTER_OTLP_HEADERS=Authorization=Basic <base64(instance_id:api_token)>`

Keep provider-specific details (endpoints, auth) in env vars so swapping backends is configuration-only.
