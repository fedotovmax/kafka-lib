package events

import (
	"context"
	"fmt"

	"github.com/fedotovmax/kafka-lib/adapters"
)

const removeEventReserveQuery = "update events set reserved_to = null where id = $1;"

func (p *postgres) RemoveEventReserve(ctx context.Context, id string) error {

	const op = "adapter.db.postgres.RemoveEventReserve"

	tx := p.ex.ExtractTx(ctx)

	_, err := tx.Exec(ctx, removeEventReserveQuery, id)

	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
	}

	return nil
}
