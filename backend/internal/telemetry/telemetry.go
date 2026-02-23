package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type Config struct {
	ServiceName string
	Environment string
	Version     string

	TracesExporter string // "console", "otlp", "none"
	OTLPEndpoint   string
	OTLPHeaders    map[string]string
}

type ShutdownFunc func(context.Context) error

func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	serviceName := strings.TrimSpace(cfg.ServiceName)
	if serviceName == "" {
		serviceName = "backend"
	}

	baseRes := resource.Default()
	res, err := resource.Merge(
		baseRes,
		resource.NewWithAttributes(
			baseRes.SchemaURL(),
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(strings.TrimSpace(cfg.Version)),
			semconv.DeploymentEnvironment(strings.TrimSpace(cfg.Environment)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("build otel resource: %w", err)
	}

	traceExporter, err := buildTraceExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if traceExporter == nil {
		otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithResource(res)))
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return func(context.Context) error { return nil }, nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

func buildTraceExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.TracesExporter)) {
	case "", "console":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}
		return exp, nil
	case "otlp":
		endpoint := strings.TrimSpace(cfg.OTLPEndpoint)
		if endpoint == "" {
			endpoint = "http://localhost:4318"
		}

		opts := []otlptracehttp.Option{}
		switch {
		case strings.HasPrefix(endpoint, "https://"):
			opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
		case strings.HasPrefix(endpoint, "http://"):
			opts = append(opts, otlptracehttp.WithEndpointURL(endpoint), otlptracehttp.WithInsecure())
		default:
			opts = append(opts, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())
		}

		if len(cfg.OTLPHeaders) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
		}

		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("create otlp http trace exporter: %w", err)
		}
		return exp, nil
	case "none", "noop", "disabled", "off":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown OTEL_TRACES_EXPORTER %q (expected console|otlp|none)", cfg.TracesExporter)
	}
}

func ParseOTLPHeaders(raw string) map[string]string {
	headers := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}

		headers[key] = value
	}

	if len(headers) == 0 {
		return nil
	}
	return headers
}
