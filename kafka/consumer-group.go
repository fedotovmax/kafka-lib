package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

var ErrConsumerHandlerClosedByCtx = errors.New("handler closed by context")

type consumerGroup struct {
	consumer  sarama.ConsumerGroup
	handler   sarama.ConsumerGroupHandler
	log       *slog.Logger
	topics    []string
	sleep     time.Duration
	ctx       context.Context
	stop      context.CancelFunc
	isStopped chan struct{}
}

type ConsumerGroup interface {
	Stop(context.Context) error
	Start()
}

func NewConsumerGroup(cgcfg *ConsumerGroupConfig, log *slog.Logger, handler sarama.ConsumerGroupHandler) (ConsumerGroup, error) {
	const op = "queues.kafka.NewConsumerGroup"

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_1_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.IsolationLevel = sarama.ReadUncommitted
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.AutoCommit.Enable = cgcfg.AutoCommit

	cg, err := sarama.NewConsumerGroup(cgcfg.Brokers, cgcfg.GroupID, cfg)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &consumerGroup{
		consumer:  cg,
		handler:   handler,
		topics:    cgcfg.Topics,
		log:       log,
		sleep:     cgcfg.SleepAfterRebalance,
		ctx:       ctx,
		stop:      cancel,
		isStopped: make(chan struct{}),
	}, nil
}

func (cg *consumerGroup) readErrors(wg *sync.WaitGroup) {

	const op = "queues.kafka.consumer-group.readErrors"

	l := cg.log.With(slog.String("op", op))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-cg.ctx.Done():
				l.Info("context done, exit consumer group errors-reader")
				return
			case err, ok := <-cg.consumer.Errors():
				if !ok {
					l.Info("consumer group errors channel was closed, exit reading errors")
					return
				}
				if err != nil {
					l.Error("consumer group error", slog.String("error", err.Error()))
				}
			}
		}
	}()
}

func (cg *consumerGroup) consume(wg *sync.WaitGroup) {
	const op = "queues.kafka.consumer-group.consume"

	l := cg.log.With(slog.String("op", op))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if cg.ctx.Err() != nil {
				l.Info("context done, exit consuming")
				return
			}

			err := cg.consumer.Consume(cg.ctx, cg.topics, cg.handler)

			if err != nil {
				l.Error(err.Error())
				time.Sleep(cg.sleep)
			}
		}
	}()
}

func (cg *consumerGroup) Start() {

	wg := &sync.WaitGroup{}

	cg.readErrors(wg)
	cg.consume(wg)

	go func() {
		wg.Wait()
		close(cg.isStopped)
	}()

}

func (cg *consumerGroup) Stop(ctx context.Context) error {

	const op = "queues.kafka.consumer-group.Stop"

	done := make(chan error, 1)

	cg.stop()

	go func() {
		<-cg.isStopped
		err := cg.consumer.Close()
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {

			if errors.Is(err, ErrConsumerHandlerClosedByCtx) {
				return nil
			}

			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}
}
