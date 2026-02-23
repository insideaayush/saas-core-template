package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type PostHogClient struct {
	apiKey string
	host   string
	client *http.Client
}

func NewPostHog(apiKey string, host string) *PostHogClient {
	key := strings.TrimSpace(apiKey)
	base := strings.TrimRight(strings.TrimSpace(host), "/")
	if base == "" {
		base = "https://app.posthog.com"
	}

	return &PostHogClient{
		apiKey: key,
		host:   base,
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *PostHogClient) Track(ctx context.Context, event Event) {
	if strings.TrimSpace(c.apiKey) == "" {
		return
	}
	name := strings.TrimSpace(event.Name)
	if name == "" {
		return
	}

	distinctID := strings.TrimSpace(event.DistinctID)
	if distinctID == "" {
		distinctID = "anonymous"
	}

	body, err := json.Marshal(map[string]any{
		"api_key":     c.apiKey,
		"event":       name,
		"distinct_id": distinctID,
		"properties":  event.Properties,
	})
	if err != nil {
		slog.Debug("failed to encode posthog event", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+"/capture/", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Debug("posthog track failed", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Debug("posthog track non-2xx", "status", resp.StatusCode, "event", name)
		return
	}
}

func ProviderFromEnv(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "console":
		return "console", nil
	case "posthog":
		return "posthog", nil
	case "none", "noop", "disabled", "off":
		return "none", nil
	default:
		return "", fmt.Errorf("unknown ANALYTICS_PROVIDER %q (expected console|posthog|none)", value)
	}
}
