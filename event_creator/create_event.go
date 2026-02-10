package eventcreator

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/outbox"
)

func (u *creator) CreateEvent(ctx context.Context, d *outbox.CreateEvent) (string, error) {
	const op = "event_creator.ConfirmFailedEvent"

	res, err := u.storage.CreateEvent(ctx, d)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}
