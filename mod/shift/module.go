package shift

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

type Module struct {
	node   node.Node
	log    *log.Logger
	config Config
}

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Service{Module: m},
	).Run(ctx)
}
