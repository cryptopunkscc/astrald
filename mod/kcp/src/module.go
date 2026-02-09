package kcp

import (
	"context"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/sig"
)

type Settings struct {
	Listen *tree.Value[*astral.Bool] `tree:"listen"`
	Dial   *tree.Value[*astral.Bool] `tree:"dial"`
}

// Module represents the KCP module and implements the exonet.Dialer interface.
type Module struct {
	Deps
	config   Config
	settings Settings

	node   astral.Node
	assets assets.Assets
	log    *log.Logger
	ctx    *astral.Context
	ops    ops.Set

	mu                    sync.Mutex
	configEndpoints       []exonet.Endpoint
	ephemeralListeners    sig.Map[astral.Uint16, exonet.EphemeralListener]
	ephemeralPortMappings sig.Map[astral.String8, astral.Uint16]

	server sig.Switch
}

func (mod *Module) GetOpSet() *ops.Set {
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

	if err := mod.loadSettings(ctx); err != nil {
		return err
	}

	go func() {
		for v := range mod.settings.Listen.Follow(ctx) {
			mod.server.Set(ctx, v == nil || bool(*v), mod.startServer)
		}
	}()

	<-ctx.Done()

	return nil
}

func (mod *Module) loadSettings(ctx *astral.Context) error {
	if mod.config.Dial != nil {
		val := astral.Bool(*mod.config.Dial)
		if err := mod.settings.Dial.Set(ctx, &val); err != nil {
			return err
		}
	}

	if mod.config.Listen != nil {
		val := astral.Bool(*mod.config.Listen)
		if err := mod.settings.Listen.Set(ctx, &val); err != nil {
			return err
		}
	}

	return nil
}

func (mod *Module) acceptAll(ctx context.Context, conn exonet.Conn) (shouldStop bool, err error) {
	err = mod.Nodes.AcceptInboundLink(ctx, conn)
	if err != nil {
		return false, err
	}

	return false, nil
}
