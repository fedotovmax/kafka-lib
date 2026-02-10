package eventcreator

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/outbox"
)

func (u *creator) ConfirmFailedEvent(ctx context.Context, ev outbox.FailedEvent) error {
	const op = "event_creator.ConfirmFailedEvent"

	err := u.txm.Wrap(ctx, func(txCtx context.Context) error {

		err := u.storage.RemoveEventReserve(txCtx, ev.GetID())

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
