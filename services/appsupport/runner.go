package appsupport

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
)

type Runner struct {
}

func (runner *Runner) Run(ctx context.Context, core api.Core) error {
	unix := &AppSupport{
		network: core.Network(),
	}

	return unix.Run(ctx)
}

func init() {
	_ = node.RegisterService("unix", &Runner{})
}
