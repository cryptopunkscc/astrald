package shell

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

var _ scheduler.Action = &LogAction{}

type LogAction struct {
	mod     *Module
	message string
}

func (l LogAction) String() string {
	return "shell.log_action"
}

func (l LogAction) Run(ctx *astral.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 5):
	}

	l.mod.log.Log(l.message)
	return nil
}
