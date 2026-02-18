package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env         string
	Version     string
	Port        string
	DatabaseURL string
	RedisURL    string

	ClerkSecretKey         string
	ClerkAPIURL            string
	StripeSecretKey        string
	StripeWebhookSecret    string
	StripeAPIURL           string
	StripePriceProMonthly  string
	StripePriceTeamMonthly string
	AppBaseURL             string
}

func Load() (Config, error) {
	cfg := Config{
		Env:         getEnv("APP_ENV", "development"),
		Version:     getEnv("APP_VERSION", "dev"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),

		ClerkSecretKey:         os.Getenv("CLERK_SECRET_KEY"),
		ClerkAPIURL:            getEnv("CLERK_API_URL", ""),
		StripeSecretKey:        os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:    os.Getenv("STRIPE_WEBHOOK_SECRET"),
		StripeAPIURL:           getEnv("STRIPE_API_URL", ""),
		StripePriceProMonthly:  getEnv("STRIPE_PRICE_PRO_MONTHLY", ""),
		StripePriceTeamMonthly: getEnv("STRIPE_PRICE_TEAM_MONTHLY", ""),
		AppBaseURL:             getEnv("APP_BASE_URL", "http://localhost:3000"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
