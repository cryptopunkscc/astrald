package kcp

import (
	"context"
	"fmt"
	"sync"

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

	mu                    sync.Mutex
	configEndpoints       []exonet.Endpoint
	ephemeralListeners    sig.Map[astral.Uint16, exonet.EphemeralListener]
	ephemeralPortMappings sig.Map[astral.String, astral.Uint16]
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) String() string {
	return kcp.ModuleName
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx
	listenPort := astral.Uint16(mod.config.ListenPort)

	err := tasks.Group(NewServer(mod, listenPort, mod.acceptAll)).Run(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

// CreateEphemeralListener creates an ephemeral KCP endpoint which will start a server that listens on the specified local endpoint and adds it to the ephemeralListeners set.
func (mod *Module) CreateEphemeralListener(port astral.Uint16) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	if _, ok := mod.ephemeralListeners.Get(port); ok {
		return fmt.Errorf("%w: port %d", kcp.ErrEphemeralListenerExists, port)
	}

	kcpServer := NewServer(mod, port, mod.acceptAll)
	go func() {
		err := kcpServer.Run(mod.ctx)
		if err != nil {
			mod.log.Error("ephemeral listener error: %v", err)
		}

		mod.ephemeralListeners.Delete(port)
	}()

	mod.ephemeralListeners.Set(port, kcpServer)
	return nil
}

func (mod *Module) acceptAll(ctx context.Context, conn exonet.Conn) (shouldStop bool, err error) {
	err = mod.Nodes.Accept(ctx, conn)
	if err != nil {
		return false, err
	}

	return false, nil
}
