package eventsender

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/outbox"
)

func (u *eventsender) CreateEvent(ctx context.Context, d *outbox.CreateEvent) (string, error) {
	const op = "usecase.events.ConfirmFailedEvent"

	res, err := u.storage.CreateEvent(ctx, d)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}
