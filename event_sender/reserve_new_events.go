package eventsender

import (
	"context"
	"fmt"
	"time"

	"github.com/fedotovmax/goutils/sliceutils"
	"github.com/fedotovmax/kafka-lib/outbox"
)

func (u *eventsender) ReserveNewEvents(ctx context.Context, limit int, reserveDuration time.Duration) ([]outbox.Event, error) {

	const op = "usecase.events.ReserveNewEvents"

	var events []*outbox.EventModel

	err := u.txm.Wrap(ctx, func(txCtx context.Context) error {
		var err error
		events, err = u.storage.FindNewAndNotReservedEvents(txCtx, limit)

		if err != nil {
			return err
		}

		eventsIds := make([]string, len(events))

		for i := 0; i < len(events); i++ {
			eventsIds[i] = events[i].ID
		}

		err = u.storage.SetEventsReservedToByIDs(txCtx, eventsIds, reserveDuration)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sliceutils.SliceToSliceInterface[*outbox.EventModel, outbox.Event](events), nil
}
