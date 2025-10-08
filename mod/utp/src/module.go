package utp

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
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
	publicEndpoints []exonet.Endpoint
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

func (mod *Module) localEndpoints() (list []exonet.Endpoint) {
	ips, err := mod.IP.LocalIPs()
	if err != nil {
		return
	}

	for _, tip := range ips {
		if tip.IsLoopback() {
			continue
		}

		if tip.IsGlobalUnicast() || tip.IsPrivate() {
			e := tcp.Endpoint{
				IP:   tip,
				Port: astral.Uint16(uint16(mod.config.ListenPort)),
			}

			list = append(list, &e)
		}

	}
	return
}
