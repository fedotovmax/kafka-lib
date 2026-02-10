package outbox

import (
	"log/slog"
	"sync"
)

func (a *Outbox) successesMonitoring(wg *sync.WaitGroup) {
	const op = "outbox.successesMonitoring"

	log := a.log.With(slog.String("op", op))

	eventsSuccesses := a.kafka.GetSuccesses(a.ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.ctx.Done():
				log.Info("monitoring [successes] stopped: ctx closed")
				return
			case event, ok := <-eventsSuccesses:
				if !ok {
					log.Info("monitoring [successes] stopped: channel closed")
					return
				}
				err := a.confirm(event)
				if err != nil {
					log.Error("error when confirm event, but event is sended",
						slog.String("error", err.Error()))
					continue
				}
				log.Info("event sended", slog.String("event_id", event.ID))
			}
		}
	}()
}
