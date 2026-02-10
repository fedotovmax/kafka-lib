package outbox

import (
	"context"
	"log/slog"
)

func (a *Outbox) process() {

	const op = "outbox.process"

	log := a.log.With(slog.String("op", op))

	queriesCtx, cancelQueriesCtx := context.WithTimeout(a.ctx, a.cfg.ProcessTimeout)
	defer cancelQueriesCtx()

	events, err := a.adapter.ReserveNewEvents(queriesCtx, a.cfg.Limit, a.cfg.ReserveDuration)

	if err != nil {
		log.Error("error when processing", slog.String("error", err.Error()))
		return
	}

	if len(events) == 0 {
		log.Debug("skip processing, no new events")
		return
	}

	publishCtx, cancelPublishCtx := context.WithTimeout(a.ctx, a.cfg.ProcessTimeout)
	defer cancelPublishCtx()

	for _, event := range events {
		err := a.kafka.Publish(publishCtx, event)
		if err != nil {
			log.Error("publish error", slog.String("error", err.Error()))
		}
	}
}
