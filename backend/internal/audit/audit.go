package audit

import "context"

type Event struct {
	OrganizationID string
	UserID         string
	Action         string
	Data           map[string]any
}

type Recorder interface {
	Record(ctx context.Context, event Event) error
}

type Reader interface {
	ListByOrganization(ctx context.Context, organizationID string, limit int) ([]EventRecord, error)
}

type EventRecord struct {
	ID             string         `json:"id"`
	OrganizationID string         `json:"organizationId"`
	UserID         string         `json:"userId"`
	Action         string         `json:"action"`
	Data           map[string]any `json:"data"`
	CreatedAt      string         `json:"createdAt"`
}

type noopRecorder struct{}

func NewNoop() Recorder { return &noopRecorder{} }

func (r *noopRecorder) Record(context.Context, Event) error { return nil }
