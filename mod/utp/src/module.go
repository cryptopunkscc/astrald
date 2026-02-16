package utp

import (
	"context"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/utp"
	"github.com/cryptopunkscc/astrald/tasks"
)

// Module represents the UDP module and implements the exonet.Dialer interface.
type Module struct {
	Deps
	config          Config // Configuration for the module
	node            astral.Node
	assets          assets.Assets
	log             *log.Logger
	ctx             context.Context
	configEndpoints []exonet.Endpoint
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	err := tasks.Group(NewServer(mod)).Run(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) endpoints() (list []*nodes.ResolvedEndpoint) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		e := &utp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, nodes.NewResolvedEndpoint(e, 7*24*time.Hour))
	}

	return list
}
