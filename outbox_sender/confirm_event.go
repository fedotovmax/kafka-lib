package outboxsender

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/outbox"
)

func (u *eventsender) ConfirmEvent(ctx context.Context, ev outbox.SuccessEvent) error {

	const op = "usecase.events.ConfirmEvent"

	err := u.txm.Wrap(ctx, func(txCtx context.Context) error {

		err := u.storage.RemoveEventReserve(txCtx, ev.GetID())

		if err != nil {
			return err
		}

		err = u.storage.SetEventStatusDone(txCtx, ev.GetID())

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
