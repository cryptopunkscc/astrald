package nat

import (
	"sync"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
)

// Ensure Module struct implements the public nat.Module interface
var _ nat.Module = &Module{}

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

	enabled atomic.Bool
	cond    *sync.Cond
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	<-ctx.Done()

	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) SetEnabled(enabled bool) {
	if mod.enabled.Swap(enabled) != enabled {
		mod.cond.Broadcast()
	}
}

func (mod *Module) String() string {
	return nat.ModuleName
}

func (mod *Module) addTraversedPair(
	traversedEndpointPair nat.TraversedPortPair,
	initiatedByLocal bool,
) {
	mod.log.Info("added NAT traversed Pair: %v (%v) <-> %v (%v) nonce=%v",
		traversedEndpointPair.PeerA.Identity,
		traversedEndpointPair.PeerA.Endpoint,
		traversedEndpointPair.PeerB.Identity,
		traversedEndpointPair.PeerB.Endpoint,
		traversedEndpointPair.Nonce,
	)

	pair, err := NewPair(traversedEndpointPair, mod.ctx.Identity(), initiatedByLocal, WithOnPairExpire(func(p *Pair) {
		mod.log.Info("expired NAT traversed Pair: %v (%v) <-> %v (%v) nonce=%v",
			p.PeerA.Identity,
			p.PeerA.Endpoint,
			p.PeerB.Identity,
			p.PeerB.Endpoint,
			p.Nonce,
		)

		mod.pool.Remove(p.Nonce)
	}))
	if err != nil {
		mod.log.Error("error while creating pair: %v", err)
		return
	}

	err = pair.StartKeepAlive(mod.ctx)
	if err != nil {
		mod.log.Error("error starting pair keep-alive: %v", err)
	}

	err = mod.pool.Add(pair)
	if err != nil {
		mod.log.Error("error while adding Pair to pool: %v", err)
	}
}

func (mod *Module) traversedPairs() []*Pair {
	return mod.pool.pairs.Values()
}
