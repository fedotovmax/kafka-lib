package outbox

import (
	"context"
	"fmt"
)

func (a *Outbox) confirm(ev SuccessEvent) error {

	const op = "outbox.confirm"

	queriesCtx, cancelQueriesCtx := context.WithTimeout(a.ctx, a.cfg.ProcessTimeout)
	defer cancelQueriesCtx()

	err := a.adapter.ConfirmEvent(queriesCtx, ev)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}
