package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"saas-core-template/backend/internal/config"
	"saas-core-template/backend/internal/db"
	"saas-core-template/backend/internal/email"
	"saas-core-template/backend/internal/errorreporting"
	"saas-core-template/backend/internal/jobs"
	"saas-core-template/backend/internal/telemetry"
)

const workerName = "saas-core-template-worker"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if !cfg.JobsEnabled {
		slog.Info("jobs disabled; exiting")
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	shutdownTelemetry, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    defaultString(cfg.ServiceName, workerName),
		Environment:    cfg.Env,
		Version:        cfg.Version,
		TracesExporter: cfg.OtelTracesExporter,
		OTLPEndpoint:   cfg.OtelOTLPEndpoint,
		OTLPHeaders:    telemetry.ParseOTLPHeaders(cfg.OtelOTLPHeadersRaw),
	})
	if err != nil {
		slog.Error("failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	reporter, err := errorreporting.New(ctx, errorreporting.Config{
		Provider:    cfg.ErrorReportingProvider,
		DSN:         cfg.SentryDSN,
		Environment: cfg.SentryEnvironment,
		Release:     cfg.Version,
	})
	if err != nil {
		slog.Error("failed to initialize error reporting", "error", err)
		os.Exit(1)
	}
	defer func() { _ = reporter.Shutdown(context.Background()) }()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	sender := buildEmailSender(cfg)
	claimer := jobs.NewClaimer(pool, jobs.ClaimerConfig{
		WorkerID: cfg.JobsWorkerID,
		LockTTL:  5 * time.Minute,
	})

	slog.Info("worker started", "name", workerName, "worker_id", cfg.JobsWorkerID, "poll", cfg.JobsPollInterval.String())

	ticker := time.NewTicker(cfg.JobsPollInterval)
	defer ticker.Stop()

	from := defaultString(cfg.EmailFrom, "local@example.com")

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down")
			return
		case <-ticker.C:
			if err := runOnce(ctx, claimer, sender, from); err != nil && !errors.Is(err, context.Canceled) {
				errorreporting.Capture(ctx, reporter, err, map[string]string{"component": "worker"})
			}
		}
	}
}

func runOnce(ctx context.Context, claimer *jobs.Claimer, sender email.Sender, from string) error {
	job, err := claimer.ClaimNext(ctx)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	switch job.Type {
	case "send_email":
		var payload struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Text    string `json:"text"`
			HTML    string `json:"html"`
		}
		if err := json.Unmarshal(job.PayloadJSON, &payload); err != nil {
			_ = claimer.Fail(ctx, jobs.FailureInput{JobID: job.ID, Attempts: job.Attempts, MaxAttempts: job.MaxAttempts, Err: fmt.Errorf("decode payload: %w", err)})
			return nil
		}

		msg := email.Message{
			To:      payload.To,
			From:    from,
			Subject: payload.Subject,
			Text:    payload.Text,
			HTML:    payload.HTML,
		}

		if err := sender.Send(ctx, msg); err != nil {
			_ = claimer.Fail(ctx, jobs.FailureInput{JobID: job.ID, Attempts: job.Attempts, MaxAttempts: job.MaxAttempts, Err: err})
			return nil
		}
		return claimer.Complete(ctx, job.ID)
	default:
		_ = claimer.Fail(ctx, jobs.FailureInput{JobID: job.ID, Attempts: job.Attempts, MaxAttempts: job.MaxAttempts, Err: fmt.Errorf("unknown job type %q", job.Type)})
		return nil
	}
}

func buildEmailSender(cfg config.Config) email.Sender {
	switch cfg.EmailProvider {
	case "", "console":
		return email.NewConsole()
	case "none", "noop", "off", "disabled":
		return email.NewNoop()
	case "resend":
		return email.NewResend(cfg.ResendAPIKey)
	default:
		return email.NewConsole()
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
