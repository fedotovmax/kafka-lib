package eventsender

import (
	"context"
	"time"

	"github.com/fedotovmax/kafka-lib/outbox"
	"github.com/fedotovmax/pgxtx"
)

type Storage interface {
	SetEventStatusDone(ctx context.Context, id string) error
	SetEventsReservedToByIDs(ctx context.Context, ids []string, dur time.Duration) error
	RemoveEventReserve(ctx context.Context, id string) error
	CreateEvent(ctx context.Context, d *outbox.CreateEvent) (string, error)
	FindNewAndNotReservedEvents(ctx context.Context, limit int) ([]*outbox.EventModel, error)
}

type eventsender struct {
	storage Storage
	txm     pgxtx.Manager
}

func New(storage Storage, txm pgxtx.Manager) *eventsender {
	return &eventsender{storage: storage, txm: txm}
}
