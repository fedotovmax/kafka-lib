package eventspostgres

import (
	"context"
	"fmt"
	"time"

	"github.com/fedotovmax/kafka-lib/adapters"
)

const setEventsReservedToByIDsQuery = "update events set reserved_to = $1 where id = ANY ($2);"

func (p *postgres) SetEventsReservedToByIDs(ctx context.Context, ids []string, dur time.Duration) error {

	const op = "adapter.db.postgres.SetEventsReservedToByIDs"

	reservedTo := time.Now().Add(dur).UTC()

	tx := p.ex.ExtractTx(ctx)

	_, err := tx.Exec(ctx, setEventsReservedToByIDsQuery, reservedTo, ids)

	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, adapters.ErrInternal, err)
	}

	return nil

}
