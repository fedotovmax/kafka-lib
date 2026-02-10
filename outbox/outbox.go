package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/fedotovmax/kafka-lib/kafka"
)

type Outbox struct {
	kafka     *produceKafka
	log       *slog.Logger
	cfg       *Config
	adapter   Adapter
	ctx       context.Context
	stop      context.CancelFunc
	isStopped chan struct{}
	inProcess int32
}

// Limit = 50, ProcessTimeout = 360ms -> For kafka flush: MaxMessages = 12-25, Frequency = 90-180ms;
// Limit = 200, ProcessTimeout = 530ms -> For kafka flush: MaxMessages = 50-100, Frequency = 130-265ms;
// Limit = 500, ProcessTimeout = 720ms -> For kafka flush: MaxMessages = 125-250, Frequency = 180-360ms;
func New(l *slog.Logger, p kafka.Producer, ad Adapter, cfg *Config) (*Outbox, error) {

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
