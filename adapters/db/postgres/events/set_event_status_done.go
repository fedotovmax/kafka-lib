package events

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/adapters"
	"github.com/fedotovmax/kafka-lib/outbox"
)

const setEventStatusDoneQuery = "update events set status = $1 where id = $2;"

func (p *postgres) SetEventStatusDone(ctx context.Context, id string) error {
	const op = "adapter.db.postgres.SetEventStatusDone"

	tx := p.ex.ExtractTx(ctx)

	_, err := tx.Exec(ctx, setEventStatusDoneQuery, outbox.EventStatusDone, id)

	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
	}

	return nil
}
