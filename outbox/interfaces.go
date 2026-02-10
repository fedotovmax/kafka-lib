package outbox

import (
	"context"
	"encoding/json"
	"time"
)

type FailedEvent interface {
	GetID() string
	GetType() string
	GetError() error
}

type SuccessEvent interface {
	GetID() string
	GetType() string
}

type Event interface {
	GetID() string
	GetAggregateID() string
	GetTopic() string
	GetType() string
	GetPayload() json.RawMessage
}

type Creator interface {
	CreateEvent(ctx context.Context, d *CreateEvent) (string, error)
}

type Adapter interface {
	ConfirmFailedEvent(ctx context.Context, ev FailedEvent) error
	ConfirmEvent(ctx context.Context, ev SuccessEvent) error
	ReserveNewEvents(ctx context.Context, limit int, reserveDuration time.Duration) ([]Event, error)
}
