package outbox

import (
	"context"
	"fmt"
)

func (a *Outbox) fail(ev *failedEvent) error {
	const op = "outbox.fail"

	queriesCtx, cancelQueriesCtx := context.WithTimeout(a.ctx, a.cfg.ProcessTimeout)
	defer cancelQueriesCtx()

	err := a.adapter.ConfirmFailedEvent(queriesCtx, ev)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
