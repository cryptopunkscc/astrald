package kcp

import (
	"context"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
)

// Module represents the KCP module and implements the exonet.Dialer interface.
type Module struct {
	Deps
	config Config
	node   astral.Node
	assets assets.Assets
	log    *log.Logger
	ctx    *astral.Context
	ops    shell.Scope

	configEndpoints       []exonet.Endpoint
	ephemeralListeners    sig.Map[int, exonet.EphemeralListener]
	ephemeralPortMappings sig.Map[int, astral.Uint16]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	err := tasks.Group(NewServer(mod, mod.config.ListenPort, func(ctx context.Context, conn exonet.Conn) (shouldStop bool, err error) {
		err = mod.Nodes.Accept(ctx, conn)
		if err != nil {
			return true, err
		}

		return false, nil
	})).Run(ctx)
	if err != nil {
		return err
	}
	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) endpoints() (list []exonet.Endpoint) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		e := kcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, &e)
	}

	for port, _ := range mod.ephemeralListeners.Clone() {
		for _, tip := range ips {
			e := kcp.Endpoint{
				IP:   tip,
				Port: astral.Uint16(port),
			}

			list = append(list, &e)
		}
	}

	return list
}

// CreateEphemeralListener creates an ephemeral KCP endpoint which will start a server that listens on the specified local endpoint and adds it to the ephemeralListeners set.
func (mod *Module) CreateEphemeralListener(port int) (err error) {
	acceptAll := func(ctx context.Context, conn exonet.Conn) (shouldStop bool, err error) {
		err = mod.Nodes.Accept(ctx, conn)
		if err != nil {
			return true, err
		}

		return false, nil
	}

	kcpServer := NewServer(mod, port, acceptAll)
	go func() {
		err := kcpServer.Run(mod.ctx)
		if err != nil {
			mod.log.Error("ephemeral listener error: %v", err)
		}

		mod.ephemeralListeners.Delete(port)
	}()

	_, ok := mod.ephemeralListeners.Set(port, kcpServer)
	if !ok {
		// NOTE: such server should never start in first place, if we could not add it to map
		kcpServer.Close()
		return fmt.Errorf("failed to add ephemeral listener for %d", port)
	}

	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return kcp.ModuleName
}
