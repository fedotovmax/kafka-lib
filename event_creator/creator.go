package eventcreator

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

type creator struct {
	storage Storage
	txm     pgxtx.Manager
}

func New(storage Storage, txm pgxtx.Manager) *creator {
	return &creator{storage: storage, txm: txm}
}
