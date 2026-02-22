package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"saas-core-template/backend/internal/analytics"
	"saas-core-template/backend/internal/api"
	"saas-core-template/backend/internal/auth"
	"saas-core-template/backend/internal/billing"
	"saas-core-template/backend/internal/cache"
	"saas-core-template/backend/internal/config"
	"saas-core-template/backend/internal/db"
	"saas-core-template/backend/internal/errorreporting"
	"saas-core-template/backend/internal/telemetry"
)

const appName = "saas-core-template-api"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	analyticsProvider, err := analytics.ProviderFromEnv(cfg.AnalyticsProvider)
	if err != nil {
		slog.Error("failed to parse analytics provider", "error", err)
		os.Exit(1)
	}

	var analyticsClient analytics.Client
	switch analyticsProvider {
	case "none":
		analyticsClient = analytics.NewNoop()
	case "posthog":
		analyticsClient = analytics.NewPostHog(cfg.PostHogProjectKey, cfg.PostHogHost)
	default:
		analyticsClient = analytics.NewConsole()
	}

	shutdownTelemetry, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    cfg.ServiceName,
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
	defer func() {
		_ = shutdownTelemetry(context.Background())
	}()

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
	defer func() {
		_ = reporter.Shutdown(context.Background())
	}()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisClient, err := cache.Connect(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	var authService *auth.Service
	if cfg.ClerkSecretKey != "" {
		authProvider := auth.NewClerkProvider(cfg.ClerkSecretKey, cfg.ClerkAPIURL)
		authService = auth.NewService(authProvider, pool)
	}

	var billingService *billing.Service
	if cfg.StripeSecretKey != "" {
		stripeProvider := billing.NewStripeProvider(cfg.StripeSecretKey, cfg.StripeAPIURL)
		billingService = billing.NewService(stripeProvider, pool, cfg.StripeWebhookSecret)
		if err := billingService.EnsureDefaultPlans(ctx, billing.PlanCatalog{
			ProMonthly:  cfg.StripePriceProMonthly,
			TeamMonthly: cfg.StripePriceTeamMonthly,
		}); err != nil {
			slog.Warn("failed to ensure default billing plans", "error", err)
		}
	}

	apiServer := api.NewServer(
		appName,
		cfg.Env,
		cfg.Version,
		pool,
		redisClient,
		api.WithAuthService(authService),
		api.WithBillingService(billingService),
		api.WithAppBaseURL(cfg.AppBaseURL),
		api.WithAnalytics(analyticsClient),
	)

	baseHandler := apiServer.Handler()
	baseHandler = errorreporting.NewMiddleware(reporter).Wrap(baseHandler)
	baseHandler = otelhttp.NewHandler(baseHandler, "http")

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           baseHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("api server started", "addr", httpServer.Addr, "env", cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server terminated unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	waitForShutdown(httpServer)
}

func waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
