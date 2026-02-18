package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultStripeAPIBaseURL = "https://api.stripe.com/v1"

type StripeProvider struct {
	secretKey string
	apiBase   string
	client    *http.Client
}

func NewStripeProvider(secretKey string, apiBase string) *StripeProvider {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		base = defaultStripeAPIBaseURL
	}

	return &StripeProvider{
		secretKey: strings.TrimSpace(secretKey),
		apiBase:   strings.TrimRight(base, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, input CheckoutSessionInput) (CheckoutSession, error) {
	values := url.Values{}
	values.Set("mode", "subscription")
	values.Set("line_items[0][price]", input.PriceID)
	values.Set("line_items[0][quantity]", "1")
	values.Set("success_url", input.SuccessURL)
	values.Set("cancel_url", input.CancelURL)
	values.Set("allow_promotion_codes", "true")
	values.Set("metadata[organization_id]", input.OrganizationID)

	if input.CustomerID != "" {
		values.Set("customer", input.CustomerID)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.apiBase+"/checkout/sessions",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("build stripe checkout session request: %w", err)
	}

	req.SetBasicAuth(p.secretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("call stripe checkout sessions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return CheckoutSession{}, fmt.Errorf("stripe checkout session status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return CheckoutSession{}, fmt.Errorf("decode stripe checkout session: %w", err)
	}

	if parsed.URL == "" {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session missing url")
	}

	return CheckoutSession{
		ID:  parsed.ID,
		URL: parsed.URL,
	}, nil
}

func (p *StripeProvider) CreatePortalSession(ctx context.Context, input PortalSessionInput) (PortalSession, error) {
	values := url.Values{}
	values.Set("customer", input.CustomerID)
	values.Set("return_url", input.ReturnURL)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.apiBase+"/billing_portal/sessions",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return PortalSession{}, fmt.Errorf("build stripe portal session request: %w", err)
	}

	req.SetBasicAuth(p.secretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return PortalSession{}, fmt.Errorf("call stripe portal sessions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return PortalSession{}, fmt.Errorf("stripe portal session status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return PortalSession{}, fmt.Errorf("decode stripe portal session: %w", err)
	}

	if parsed.URL == "" {
		return PortalSession{}, fmt.Errorf("stripe portal session missing url")
	}

	return PortalSession{
		URL: parsed.URL,
	}, nil
}
