package outbox

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/fedotovmax/kafka-lib/kafka"
)

type Producer interface {
	GetInput() chan<- *sarama.ProducerMessage
	GetSuccesses() <-chan *sarama.ProducerMessage
	GetErrors() <-chan *sarama.ProducerError
}

type produceKafka struct {
	producer Producer

	onceSuccess sync.Once
	onceErrors  sync.Once

	successes chan *successEvent
	errors    chan *failedEvent
}

func newProduceKafka(p Producer) *produceKafka {
	return &produceKafka{
		producer:  p,
		successes: make(chan *successEvent),
		errors:    make(chan *failedEvent),
	}
}

func (p *produceKafka) Publish(ctx context.Context, ev Event) error {
	const op = "outbox.kafka.Publish"

	metadata := &messageMetadata{
		ID:   ev.GetID(),
		Type: ev.GetType(),
	}

	msg := &sarama.ProducerMessage{
		Topic: ev.GetTopic(),
		Key:   sarama.StringEncoder(ev.GetAggregateID()),
		Value: sarama.ByteEncoder(ev.GetPayload()),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte(kafka.HeaderEventID),
				Value: []byte(ev.GetID()),
			},
			{
				Key:   []byte(kafka.HeaderEventType),
				Value: []byte(ev.GetType()),
			},
		},
		Metadata: metadata,
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: event_id: %s: %w", op, ev.GetID(), ctx.Err())
	case p.producer.GetInput() <- msg:
		return nil
	}
}

func (p *produceKafka) GetSuccesses(ctx context.Context) <-chan *successEvent {
	p.onceSuccess.Do(func() {
		go func() {
			defer close(p.successes)
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-p.producer.GetSuccesses():
					if !ok {
						return
					}
					m, ok := msg.Metadata.(*messageMetadata)
					if !ok {
						continue
					}
					select {
					case <-ctx.Done():
						return
					case p.successes <- &successEvent{ID: m.ID, Type: m.Type}:
					}
				}
			}
		}()
	})

	return p.successes
}

func (p *produceKafka) GetErrors(ctx context.Context) <-chan *failedEvent {

	const op = "outbox.kafka.GetErrors"

	p.onceErrors.Do(func() {
		go func() {
			defer close(p.errors)
			for {
				select {
				case <-ctx.Done():
					return
				case produceErr, ok := <-p.producer.GetErrors():
					if !ok {
						return
					}
					m, ok := produceErr.Msg.Metadata.(*messageMetadata)
					if !ok {
						continue
					}
					select {
					case <-ctx.Done():
						return
					case p.errors <- &failedEvent{ID: m.ID, Type: m.Type, Error: fmt.Errorf("%s:%w", op,
						produceErr.Err)}:
					}
				}
			}
		}()
	})

	return p.errors
}
