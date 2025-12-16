package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fedotovmax/kafka-lib/kafka"
)

type OutboxAdapter interface {
	ConfirmFailedEvent(ctx context.Context, ev FailedEvent) error
	ConfirmEvent(ctx context.Context, ev SuccessEvent) error
	ReserveNewEvents(ctx context.Context, limit int, reserveDuration time.Duration) ([]Event, error)
}

type Outbox struct {
	kafka     *produceKafka
	adapter   OutboxAdapter
	log       *slog.Logger
	cfg       *Config
	inProcess int32
	ctx       context.Context
	stop      context.CancelFunc
	isStopped chan struct{}
}

// Limit = 50, ProcessTimeout = 360ms -> For kafka flush: MaxMessages = 12-25, Frequency = 90-180ms;
// Limit = 200, ProcessTimeout = 530ms -> For kafka flush: MaxMessages = 50-100, Frequency = 130-265ms;
// Limit = 500, ProcessTimeout = 720ms -> For kafka flush: MaxMessages = 125-250, Frequency = 180-360ms;
func New(l *slog.Logger, p kafka.Producer, ad OutboxAdapter, cfg *Config) (*Outbox, error) {

	err := validateConfig(cfg)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	kafka := newProduceKafka(p)

	return &Outbox{
		kafka:     kafka,
		log:       l,
		adapter:   ad,
		cfg:       cfg,
		ctx:       ctx,
		stop:      cancel,
		isStopped: make(chan struct{}),
	}, nil
}

func (a *Outbox) Start() {

	wg := &sync.WaitGroup{}

	a.successesMonitoring(wg)
	a.errorsMonitoring(wg)
	a.processingNewEvents(wg)

	go func() {
		wg.Wait()
		close(a.isStopped)
	}()
}

func (a *Outbox) Stop(ctx context.Context) error {
	const op = "outbox.app.Stop"
	log := a.log.With(slog.String("op", op))
	a.stop()
	select {
	case <-a.isStopped:
		log.Info("Event Processor stopped successfully")
		return nil
	case <-ctx.Done():
		log.Warn("Event Processor stopped by context")
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}
}

func (a *Outbox) successesMonitoring(wg *sync.WaitGroup) {
	const op = "outbox.app.successesMonitoring"

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

func (a *Outbox) errorsMonitoring(wg *sync.WaitGroup) {
	const op = "outbox.app.errorsMonitoring"

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

func (a *Outbox) fail(ev *failedEvent) error {
	const op = "outbox.app.fail"

	queriesCtx, cancelQueriesCtx := context.WithTimeout(a.ctx, a.cfg.ProcessTimeout)
	defer cancelQueriesCtx()

	err := a.adapter.ConfirmFailedEvent(queriesCtx, ev)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Outbox) confirm(ev SuccessEvent) error {

	const op = "outbox.app.confirm"

	queriesCtx, cancelQueriesCtx := context.WithTimeout(a.ctx, a.cfg.ProcessTimeout)
	defer cancelQueriesCtx()

	err := a.adapter.ConfirmEvent(queriesCtx, ev)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func (a *Outbox) processingNewEvents(wg *sync.WaitGroup) {
	const op = "outbox.app.processingNewEvents"

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

func (a *Outbox) process() {

	const op = "outbox.app.process"

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
