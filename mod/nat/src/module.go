package nat

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/ip"
	modnat "github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
)

// Ensure Module struct implements the public nat.Module interface
var _ modnat.Module = &Module{}

// Deps are injected by the core injector.
type Deps struct {
	Dir     dir.Module
	Objects objects.Module
	IP      ip.Module
}

// Module is the concrete implementation of the NAT module.
type Module struct {
	Deps

	ctx    *astral.Context
	node   astral.Node
	log    *log.Logger
	assets resources.Resources

	pool *PairPool
	ops  shell.Scope
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	mod.pool.RunCleanupLoop(30 * time.Second)

	<-ctx.Done()
	mod.pool.Stop()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return modnat.ModuleName
}

func (mod *Module) addTraversedPair(
	pair modnat.EndpointPair,
	initiatedByLocal bool,
) {
	mod.log.Info("added NAT traversed pair: %v (%v) <-> %v (%v) nonce=%v",
		pair.PeerA.Identity,
		pair.PeerA.Endpoint,
		pair.PeerB.Identity,
		pair.PeerB.Endpoint,
		pair.Nonce,
	)

	err := mod.pool.Add(&pair, mod.ctx.Identity(), initiatedByLocal)
	if err != nil {
		mod.log.Errorv(1, "error adding pair to pool: %v", err)
	}
}
