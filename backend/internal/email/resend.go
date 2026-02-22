package email

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

type ResendSender struct {
	apiKey string
	client *http.Client
}

func NewResend(apiKey string) *ResendSender {
	return &ResendSender{
		apiKey: strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ResendSender) Send(ctx context.Context, msg Message) error {
	if err := ValidateMessage(msg); err != nil {
		return err
	}
	if s.apiKey == "" {
		return fmt.Errorf("missing Resend API key")
	}

	payload := map[string]any{
		"from":    msg.From,
		"to":      []string{msg.To},
		"subject": msg.Subject,
	}
	if strings.TrimSpace(msg.HTML) != "" {
		payload["html"] = msg.HTML
	}
	if strings.TrimSpace(msg.Text) != "" {
		payload["text"] = msg.Text
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("call resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	return nil
}
