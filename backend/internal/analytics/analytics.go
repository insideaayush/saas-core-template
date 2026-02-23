package analytics

import "context"

type Event struct {
	Name       string
	DistinctID string
	Properties map[string]any
}

type Client interface {
	Track(ctx context.Context, event Event)
}

type noopClient struct{}

func NewNoop() Client { return &noopClient{} }

func (c *noopClient) Track(context.Context, Event) {}
