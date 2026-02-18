package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultClerkAPIBaseURL = "https://api.clerk.com"
	authProviderClerk      = "clerk"
)

type ClerkProvider struct {
	secretKey string
	apiBase   string
	client    *http.Client
}

func NewClerkProvider(secretKey string, apiBase string) *ClerkProvider {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		base = defaultClerkAPIBaseURL
	}

	return &ClerkProvider{
		secretKey: strings.TrimSpace(secretKey),
		apiBase:   strings.TrimRight(base, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *ClerkProvider) VerifyToken(ctx context.Context, token string) (VerifiedPrincipal, error) {
	userID, err := p.verifySessionToken(ctx, token)
	if err != nil {
		return VerifiedPrincipal{}, err
	}

	email, verified, err := p.fetchPrimaryEmail(ctx, userID)
	if err != nil {
		return VerifiedPrincipal{}, err
	}

	return VerifiedPrincipal{
		Provider:       authProviderClerk,
		ProviderUserID: userID,
		PrimaryEmail:   email,
		EmailVerified:  verified,
	}, nil
}

func (p *ClerkProvider) verifySessionToken(ctx context.Context, token string) (string, error) {
	payload := map[string]string{"token": token}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiBase+"/v1/sessions/verify", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build clerk verify request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call clerk verify session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("clerk verify session status %d: %s", resp.StatusCode, string(responseBody))
	}

	var parsed struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode clerk verify session: %w", err)
	}

	if strings.TrimSpace(parsed.UserID) == "" {
		return "", fmt.Errorf("clerk verify session missing user_id")
	}

	return parsed.UserID, nil
}

func (p *ClerkProvider) fetchPrimaryEmail(ctx context.Context, userID string) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiBase+"/v1/users/"+userID, nil)
	if err != nil {
		return "", false, fmt.Errorf("build clerk user request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.secretKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("call clerk user endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", false, fmt.Errorf("clerk user endpoint status %d: %s", resp.StatusCode, string(responseBody))
	}

	var parsed struct {
		PrimaryEmailAddressID string `json:"primary_email_address_id"`
		EmailAddresses        []struct {
			ID           string `json:"id"`
			EmailAddress string `json:"email_address"`
			Verification struct {
				Status string `json:"status"`
			} `json:"verification"`
		} `json:"email_addresses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", false, fmt.Errorf("decode clerk user response: %w", err)
	}

	for _, email := range parsed.EmailAddresses {
		if email.ID == parsed.PrimaryEmailAddressID {
			return email.EmailAddress, strings.EqualFold(email.Verification.Status, "verified"), nil
		}
	}

	for _, email := range parsed.EmailAddresses {
		if strings.TrimSpace(email.EmailAddress) != "" {
			return email.EmailAddress, strings.EqualFold(email.Verification.Status, "verified"), nil
		}
	}

	return "", false, nil
}
