# Outbox Package

Пакет реализует паттерн Outbox для обеспечения надежной доставки событий через Kafka

# Download

go get -u github.com/fedotovmax/outbox@{version}

## Поддерживаемые драйверы

- Kafka — [Sarama](https://github.com/IBM/sarama)

## Пример создания

```go

// logger — ваш slog.Logger

type Producer interface {
	GetInput() chan<- *sarama.ProducerMessage
	GetSuccesses() <-chan *sarama.ProducerMessage
	GetErrors() <-chan *sarama.ProducerError
}



type OutboxAdapter interface {
	ConfirmFailed(ctx context.Context, ev *FailedEvent) error
	ConfirmEvent(ctx context.Context, ev *SuccessEvent) error
	ReserveNewEvents(ctx context.Context, limit int, reserveDuration time.Duration) ([]*Event, error)
}




// config — outbox.Config

ob := outbox.New(logger, producer, adapter, config)
```

## Конструктор `New`

| Параметр | Тип            | Описание                                  |
| -------- | -------------- | ----------------------------------------- |
| `l`      | `*slog.Logger` | Логгер для вывода событий и ошибок        |
| `p`      | `Producer`     | Адаптер для публикации событий в Kafka    |
| `ad`     | `Adapter`      | Адаптер для для работы с вашем хранилищем |
| `cfg`    | `Config`       | Конфигурация Outbox                       |

## Методы

| Метод | Сигнатура | Описание |
| ------------- | ------------------------------------------------------------------ | ----------------------------------------------------------- | |
| `Start` | `Start()` | Запуск обработки событий |
| `Stop` | `Stop(ctx context.Context) error` | Остановка обработки с ожиданием завершения текущих операций |

# Consumer handler example

```go

const HeaderEventType = "event_type"
const HeaderEventID = "event_id"

func (k *kafkaController) Setup(s sarama.ConsumerGroupSession) error {

    l.Info("Setup: partitions assigned", slog.Any("claims", claims))

    // Examples for future maybe:
    // 1) Preload caches
    // k.cache.Load()

    // 2) Init workers for each partition
    // k.startWorkersForClaims(claims)

    // 3) Reset metrics
    // k.metrics.Reset()

    return nil

}

func (k *kafkaController) Cleanup(s sarama.ConsumerGroupSession) error {

    l.Info("Cleanup: partitions revoked", slog.Any("claims", claims))

    // Examples for future maybe:
    // 1) Stop workers
    // k.stopWorkers()

    // 2) Flush buffers
    // k.flush()

    // 3) Close resources

    return nil

}

func (k *kafkaController) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {

    const op = "controller.kafka_consumer.ConsumeClaim"

    l := k.log.With(slog.String("op", op))

    for {
    	select {
    	case <-s.Context().Done():
    		return fmt.Errorf("%s: %w", op, s.Context().Err())
    	case message, ok := <-c.Messages():

    		if !ok {
    			return fmt.Errorf("%s: %w", op, ErrKafkaMessagesChannelClosed)
    		}

    		var eventID string
    		var eventType string

    		for _, header := range message.Headers {
    			key := string(header.Key)
    			switch key {
    			case HeaderEventType:
    				eventType = string(header.Value)
    			case HeaderEventID:
    				eventID = string(header.Value)
    			}
    		}

    		if eventID == "" {
    			l.Error("empty event ID")
    			s.MarkMessage(message, "")
    			continue
    		}

    		if eventType == "" {
    			l.Error("empty event type")
    			s.MarkMessage(message, "")
    			continue
    		}

    		payload := message.Value

    		switch eventType {

    		case events.USER_CREATED:

    			var createdUserPayload events.UserCreatedEventPayload

    			err := json.Unmarshal(payload, &createdUserPayload)

    			if err != nil {
    				l.Error("invalid payload", logger.Err(err), slog.String("event_type", eventType))
    				s.MarkMessage(message, "")
    				continue
    			}

    			//TODO: real handle
    			l.Info(
    				"======================successfully consume message",
    				slog.Any("payload", createdUserPayload),
    				slog.Any("partition", message.Partition),
    				slog.Int64("offset", message.Offset),
    			)

    			s.MarkMessage(message, "")

    		default:
    			l.Error("invalid event type", slog.String("event_type", eventType))
    			s.MarkMessage(message, "")
    		}
    	}
    }

}
```
