package outbox

import (
	"log/slog"
	"sync"
)

func (a *Outbox) errorsMonitoring(wg *sync.WaitGroup) {
	const op = "outbox.errorsMonitoring"

	log := a.log.With(slog.String("op", op))

	eventsErrors := a.kafka.GetErrors(a.ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				log.Info("monitoring [errors] stopped: ctx closed")
				return
			case event, ok := <-eventsErrors:
				if !ok {
					log.Info("monitoring [errors] stopped: channel closed")
					return
				}
				log.Error("event send failed",
					slog.String("event_id", event.ID), slog.String("error", event.Error.Error()))

				err := a.fail(event)

				if err != nil {
					log.Error("error when confirm send fail", slog.String("error", err.Error()))
				}
			}
		}
	}()
}
