package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

type Producer interface {
	GetInput() chan<- *sarama.ProducerMessage
	GetSuccesses() <-chan *sarama.ProducerMessage
	GetErrors() <-chan *sarama.ProducerError
	Stop(context.Context) error
}

type producer struct {
	instance sarama.AsyncProducer
}

func (p *producer) GetInput() chan<- *sarama.ProducerMessage {
	return p.instance.Input()
}

func (p *producer) GetSuccesses() <-chan *sarama.ProducerMessage {
	return p.instance.Successes()
}

func (p *producer) GetErrors() <-chan *sarama.ProducerError {
	return p.instance.Errors()
}

func (p *producer) Stop(ctx context.Context) error {

	const op = "queues.kafka.producer.Close"

	done := make(chan error, 1)

	go func() {
		err := p.instance.Close()
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}
}

func NewAsyncProducer(pcfg ProducerConfig) (Producer, error) {
	const op = "queues.kafka.producer.NewAsyncProducer"

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_1_0_0
	cfg.Producer.Return.Successes = true
	cfg.Producer.Retry.Max = 5
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Flush.MaxMessages = pcfg.MaxMessages
	cfg.Producer.Flush.Frequency = pcfg.Frequency

	p, err := sarama.NewAsyncProducer(pcfg.Brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &producer{
		instance: p,
	}, nil
}
