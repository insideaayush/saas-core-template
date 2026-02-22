package errorreporting

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

type Reporter interface {
	CaptureException(ctx context.Context, err error, attrs map[string]string)
	Shutdown(ctx context.Context) error
}

type Config struct {
	Provider    string // "console", "sentry", "none"
	DSN         string
	Environment string
	Release     string
}

func New(ctx context.Context, cfg Config) (Reporter, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "console":
		return &consoleReporter{}, nil
	case "none", "noop", "disabled", "off":
		return &noopReporter{}, nil
	case "sentry":
		dsn := strings.TrimSpace(cfg.DSN)
		if dsn == "" {
			return &consoleReporter{}, nil
		}

		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      strings.TrimSpace(cfg.Environment),
			Release:          strings.TrimSpace(cfg.Release),
			AttachStacktrace: true,
		}); err != nil {
			return nil, fmt.Errorf("init sentry: %w", err)
		}

		// Confirm SDK is ready by capturing a breadcrumb-style no-op on startup.
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("component", "backend")
		})

		return &sentryReporter{}, nil
	default:
		return nil, fmt.Errorf("unknown ERROR_REPORTING_PROVIDER %q (expected console|sentry|none)", cfg.Provider)
	}
}

type noopReporter struct{}

func (r *noopReporter) CaptureException(context.Context, error, map[string]string) {}
func (r *noopReporter) Shutdown(context.Context) error                             { return nil }

type consoleReporter struct{}

func (r *consoleReporter) CaptureException(_ context.Context, err error, attrs map[string]string) {
	if err == nil {
		return
	}

	fields := []any{"error", err}
	for k, v := range attrs {
		fields = append(fields, k, v)
	}
	slog.Error("captured exception", fields...)
}

func (r *consoleReporter) Shutdown(context.Context) error { return nil }

type sentryReporter struct{}

func (r *sentryReporter) CaptureException(_ context.Context, err error, attrs map[string]string) {
	if err == nil {
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range attrs {
			scope.SetTag(k, v)
		}
		sentry.CaptureException(err)
	})
}

func (r *sentryReporter) Shutdown(ctx context.Context) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		sentry.Flush(2 * time.Second)
		return nil
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		return nil
	}

	sentry.Flush(remaining)
	return nil
}
