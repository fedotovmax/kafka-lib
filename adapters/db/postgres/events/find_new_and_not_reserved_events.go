package events

import (
	"context"
	"fmt"
	"time"

	"github.com/fedotovmax/kafka-lib/adapters"
	"github.com/fedotovmax/kafka-lib/outbox"
)

const findNewAndNotReservedEventsQuery = `select id, aggregate_id, event_topic, event_type,
	payload, status, created_at, reserved_to
	from events where status = $1 AND
	(reserved_to IS NULL OR reserved_to < $2)
	order by created_at asc
	limit $3;`

func (p *postgres) FindNewAndNotReservedEvents(ctx context.Context, limit int) ([]*outbox.EventModel, error) {

	const op = "adapter.db.postgres.FindNewAndNotReservedEvents"

	tx := p.ex.ExtractTx(ctx)

	rows, err := tx.Query(ctx, findNewAndNotReservedEventsQuery, outbox.EventStatusNew, time.Now().UTC(), limit)

	if err != nil {
		return nil, fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
	}
	defer rows.Close()

	var events []*outbox.EventModel

	for rows.Next() {

		e := &outbox.EventModel{}

		err := rows.Scan(&e.ID, &e.AggregateID, &e.Topic, &e.Type, &e.Payload,
			&e.Status, &e.CreatedAt, &e.ReservedTo)

		if err != nil {
			return nil, fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
	}

	return events, nil

}
