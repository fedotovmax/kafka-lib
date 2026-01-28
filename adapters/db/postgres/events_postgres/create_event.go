package eventspostgres

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/adapters"
	"github.com/fedotovmax/kafka-lib/outbox"
)

const createEventQuery = `insert into events (aggregate_id, event_topic, event_type, payload, created_at)
values ($1,$2,$3,$4,$5) returning id;`

func (p *postgres) CreateEvent(ctx context.Context, in *outbox.CreateEvent) (string, error) {
	const op = "adapter.db.postgres.CreateEvent"

	tx := p.ex.ExtractTx(ctx)

	row := tx.QueryRow(ctx, createEventQuery,
		in.GetAggregateID(), in.GetTopic(), in.GetType(), in.GetPayload(), in.GetCreatedAt())

	var id string

	err := row.Scan(&id)

	if err != nil {
		return "", fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
	}

	return id, nil
}
