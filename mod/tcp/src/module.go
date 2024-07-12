package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tcp.Module = &Module{}

type Module struct {
	config          Config
	node            node2.Node
	log             *log.Logger
	ctx             context.Context
	nodes           nodes.Module
	publicEndpoints []Endpoint
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(NewServer(mod)).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}
