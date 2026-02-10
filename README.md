go get github.com/fedotovmax/kafka-lib@v1.0.19
# 1. Для корректного использования в проекте нужно добавить в миграцию следующую таблицу:

```sql
create table if not exists events (
  id uuid primary key default gen_random_uuid(),
  aggregate_id varchar(100) not null,
  event_topic varchar(100) not null,
  event_type varchar(100) not null,
  payload jsonb not null,
  status varchar not null default 'new' check(status in ('new', 'done')),
  created_at timestamp not null,
  reserved_to timestamp default null
);


create index concurrently idx_events_new_unreserved_created_at
on events (created_at)
where status = 'new'
and reserved_to is null;

```

## Далее создать все сущности:

### пакет adapters/db/postgres/- создать адаптер для postgresql

### пакет outbox_sender - создать отправителя событий. Требует storage (adapter postgres)

### пакет kafka создать producer и consumer (по требованию)

### создать финальный outbox processor из пакета outbox, передав все требуемые ему завивимости.

# 2. Или сделать свою реализацию, но требуется реализовать интерфейсы:

```go
type Adapter interface {
ConfirmFailedEvent(ctx context.Context, ev FailedEvent) error
ConfirmEvent(ctx context.Context, ev SuccessEvent) error
ReserveNewEvents(ctx context.Context, limit int, reserveDuration time.Duration) ([]Event, error)
}

type Event interface {
	GetID() string
	GetAggregateID() string
	GetTopic() string
	GetType() string
	GetPayload() json.RawMessage
}

type FailedEvent interface {
	GetID() string
	GetType() string
	GetError() error
}

type SuccessEvent interface {
	GetID() string
	GetType() string
}
```
