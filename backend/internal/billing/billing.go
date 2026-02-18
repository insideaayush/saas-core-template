package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMissingWebhookSignature = errors.New("missing stripe signature")
	ErrInvalidWebhookSignature = errors.New("invalid stripe signature")
)

type Provider interface {
	CreateCheckoutSession(ctx context.Context, input CheckoutSessionInput) (CheckoutSession, error)
	CreatePortalSession(ctx context.Context, input PortalSessionInput) (PortalSession, error)
}

type Service struct {
	provider      Provider
	db            *pgxpool.Pool
	webhookSecret string
}

type PlanCatalog struct {
	Free        string
	ProMonthly  string
	TeamMonthly string
}

type CheckoutSessionInput struct {
	OrganizationID string
	CustomerID     string
	PriceID        string
	SuccessURL     string
	CancelURL      string
}

type CheckoutSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type PortalSessionInput struct {
	CustomerID string
	ReturnURL  string
}

type PortalSession struct {
	URL string `json:"url"`
}

type SubscriptionSnapshot struct {
	OrganizationID         string
	Provider               string
	ProviderCustomerID     string
	ProviderSubscriptionID string
	Status                 string
	CurrentPeriodEnd       *time.Time
}

func NewService(provider Provider, db *pgxpool.Pool, webhookSecret string) *Service {
	return &Service{
		provider:      provider,
		db:            db,
		webhookSecret: strings.TrimSpace(webhookSecret),
	}
}

func (s *Service) EnsureDefaultPlans(ctx context.Context, catalog PlanCatalog) error {
	type plan struct {
		Code            string
		DisplayName     string
		ProviderPriceID string
	}

	plans := []plan{
		{
			Code:            "pro",
			DisplayName:     "Pro",
			ProviderPriceID: strings.TrimSpace(catalog.ProMonthly),
		},
		{
			Code:            "team",
			DisplayName:     "Team",
			ProviderPriceID: strings.TrimSpace(catalog.TeamMonthly),
		},
	}

	for _, p := range plans {
		if p.ProviderPriceID == "" {
			continue
		}

		if _, err := s.db.Exec(ctx, `
			INSERT INTO plans (code, display_name, provider, provider_price_id, billing_interval, is_active)
			VALUES ($1, $2, 'stripe', $3, 'monthly', true)
			ON CONFLICT (code) DO UPDATE
			SET display_name = EXCLUDED.display_name,
			    provider = EXCLUDED.provider,
			    provider_price_id = EXCLUDED.provider_price_id,
			    billing_interval = EXCLUDED.billing_interval,
			    is_active = EXCLUDED.is_active,
			    updated_at = now()
		`, p.Code, p.DisplayName, p.ProviderPriceID); err != nil {
			return fmt.Errorf("upsert plan %s: %w", p.Code, err)
		}
	}

	return nil
}

func (s *Service) LookupPlanPriceID(ctx context.Context, planCode string) (string, error) {
	var priceID string
	if err := s.db.QueryRow(ctx, `
		SELECT provider_price_id
		FROM plans
		WHERE code = $1 AND provider = 'stripe' AND is_active = true
		LIMIT 1
	`, strings.TrimSpace(planCode)).Scan(&priceID); err != nil {
		return "", fmt.Errorf("lookup plan in db: %w", err)
	}

	return priceID, nil
}

func (s *Service) GetOrganizationCustomerID(ctx context.Context, organizationID string) (string, error) {
	var customerID string
	err := s.db.QueryRow(ctx, `
		SELECT provider_customer_id
		FROM subscriptions
		WHERE organization_id = $1
		  AND provider = 'stripe'
		  AND provider_customer_id IS NOT NULL
		ORDER BY updated_at DESC
		LIMIT 1
	`, organizationID).Scan(&customerID)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(customerID), nil
}

func (s *Service) CreateCheckoutSession(ctx context.Context, input CheckoutSessionInput) (CheckoutSession, error) {
	return s.provider.CreateCheckoutSession(ctx, input)
}

func (s *Service) CreatePortalSession(ctx context.Context, input PortalSessionInput) (PortalSession, error) {
	return s.provider.CreatePortalSession(ctx, input)
}

func (s *Service) VerifyWebhookSignature(sigHeader string, payload []byte) error {
	if s.webhookSecret == "" {
		// Allow development environments without webhook signing configured.
		return nil
	}

	sigHeader = strings.TrimSpace(sigHeader)
	if sigHeader == "" {
		return ErrMissingWebhookSignature
	}

	ts, providedSig := parseStripeSignature(sigHeader)
	if ts == "" || providedSig == "" {
		return ErrInvalidWebhookSignature
	}

	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(providedSig)) {
		return ErrInvalidWebhookSignature
	}

	return nil
}

func (s *Service) HandleWebhookEvent(ctx context.Context, payload []byte) error {
	var event struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object map[string]any `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("parse stripe event: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event.Data.Object)
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		return s.handleSubscriptionChange(ctx, event.Data.Object)
	default:
		return nil
	}
}

func (s *Service) handleCheckoutCompleted(ctx context.Context, obj map[string]any) error {
	orgID := getStringFromAnyMap(getMapFromAnyMap(obj, "metadata"), "organization_id")
	if orgID == "" {
		return nil
	}

	subscriptionID := getStringFromAnyMap(obj, "subscription")
	customerID := getStringFromAnyMap(obj, "customer")
	status := "active"
	if value := getStringFromAnyMap(obj, "status"); value != "" {
		status = value
	}

	snapshot := SubscriptionSnapshot{
		OrganizationID:         orgID,
		Provider:               "stripe",
		ProviderCustomerID:     customerID,
		ProviderSubscriptionID: subscriptionID,
		Status:                 normalizeSubscriptionStatus(status),
	}

	return s.UpsertSubscription(ctx, snapshot)
}

func (s *Service) handleSubscriptionChange(ctx context.Context, obj map[string]any) error {
	subscriptionID := getStringFromAnyMap(obj, "id")
	customerID := getStringFromAnyMap(obj, "customer")
	status := normalizeSubscriptionStatus(getStringFromAnyMap(obj, "status"))

	var periodEnd *time.Time
	if unix := getInt64FromAnyMap(obj, "current_period_end"); unix > 0 {
		t := time.Unix(unix, 0).UTC()
		periodEnd = &t
	}

	orgID, err := s.findOrganizationForStripeSubscription(ctx, subscriptionID, customerID)
	if err != nil || orgID == "" {
		return nil
	}

	snapshot := SubscriptionSnapshot{
		OrganizationID:         orgID,
		Provider:               "stripe",
		ProviderCustomerID:     customerID,
		ProviderSubscriptionID: subscriptionID,
		Status:                 status,
		CurrentPeriodEnd:       periodEnd,
	}

	return s.UpsertSubscription(ctx, snapshot)
}

func (s *Service) findOrganizationForStripeSubscription(ctx context.Context, subscriptionID string, customerID string) (string, error) {
	var orgID string

	if subscriptionID != "" {
		err := s.db.QueryRow(ctx, `
			SELECT organization_id::text
			FROM subscriptions
			WHERE provider = 'stripe' AND provider_subscription_id = $1
			ORDER BY updated_at DESC
			LIMIT 1
		`, subscriptionID).Scan(&orgID)
		if err == nil && orgID != "" {
			return orgID, nil
		}
	}

	if customerID != "" {
		err := s.db.QueryRow(ctx, `
			SELECT organization_id::text
			FROM subscriptions
			WHERE provider = 'stripe' AND provider_customer_id = $1
			ORDER BY updated_at DESC
			LIMIT 1
		`, customerID).Scan(&orgID)
		if err == nil && orgID != "" {
			return orgID, nil
		}
	}

	return "", nil
}

func (s *Service) UpsertSubscription(ctx context.Context, snapshot SubscriptionSnapshot) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO subscriptions (
			organization_id,
			provider,
			provider_customer_id,
			provider_subscription_id,
			status,
			current_period_end
		)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6)
		ON CONFLICT (provider, provider_subscription_id)
		DO UPDATE
		SET provider_customer_id = COALESCE(EXCLUDED.provider_customer_id, subscriptions.provider_customer_id),
		    status = EXCLUDED.status,
		    current_period_end = EXCLUDED.current_period_end,
		    updated_at = now()
	`, snapshot.OrganizationID, snapshot.Provider, snapshot.ProviderCustomerID, snapshot.ProviderSubscriptionID, snapshot.Status, snapshot.CurrentPeriodEnd)
	if err != nil {
		return fmt.Errorf("upsert subscription: %w", err)
	}

	return nil
}

func parseStripeSignature(sig string) (string, string) {
	parts := strings.Split(sig, ",")
	var ts, signature string

	for _, part := range parts {
		chunk := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(chunk) != 2 {
			continue
		}

		switch chunk[0] {
		case "t":
			ts = chunk[1]
		case "v1":
			signature = chunk[1]
		}
	}

	return ts, signature
}

func normalizeSubscriptionStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "active", "trialing", "past_due", "canceled", "unpaid", "incomplete", "incomplete_expired":
		return status
	default:
		return "active"
	}
}

func getStringFromAnyMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}

	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func getMapFromAnyMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}

	value, ok := m[key]
	if !ok || value == nil {
		return nil
	}

	typed, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	return typed
}

func getInt64FromAnyMap(m map[string]any, key string) int64 {
	if m == nil {
		return 0
	}

	value, ok := m[key]
	if !ok || value == nil {
		return 0
	}

	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
