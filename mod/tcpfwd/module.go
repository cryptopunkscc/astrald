package tcpfwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/tasks"
)

type Module struct {
	node   modules.Node
	assets assets.Store
	config Config
	log    *log.Logger
	ctx    context.Context
}

func (m *Module) Run(ctx context.Context) error {
	m.ctx = ctx

	var runners []tasks.Runner

	for serviceName, target := range m.config.Out {
		runners = append(runners, &ForwardOutServer{
			Module:      m,
			serviceName: serviceName,
			target:      target,
		})
	}

	for tcpAddr, target := range m.config.In {
		runners = append(runners, &ForwardInServer{
			Module:   m,
			identity: m.node.Identity(),
			tcpAddr:  tcpAddr,
			target:   target,
		})
	}

	return tasks.Group(runners...).Run(ctx)
}
