package outbox

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

func (a *Outbox) processingNewEvents(wg *sync.WaitGroup) {
	const op = "outbox.processingNewEvents"

	log := a.log.With(slog.String("op", op))

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(a.cfg.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-a.ctx.Done():
				log.Info("event processing stopped")
				return
			case <-ticker.C:
				if !atomic.CompareAndSwapInt32(&a.inProcess, 0, 1) {
					continue
				}
				a.process()
				atomic.StoreInt32(&a.inProcess, 0)
			}
		}
	}()
}
