package analytics

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
)

type ConsoleClient struct{}

func NewConsole() *ConsoleClient { return &ConsoleClient{} }

func (c *ConsoleClient) Track(_ context.Context, event Event) {
	name := strings.TrimSpace(event.Name)
	if name == "" {
		return
	}

	props := ""
	if len(event.Properties) > 0 {
		if encoded, err := json.Marshal(event.Properties); err == nil {
			props = string(encoded)
		}
	}

	slog.Info("analytics event", "name", name, "distinct_id", event.DistinctID, "properties", props)
}
