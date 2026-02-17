package nat

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/resources"
)

// Ensure Module struct implements the public nat.Module interface
var _ nat.Module = &Module{}

// Deps are injected by the core injector.
type Deps struct {
	Dir     dir.Module
	Objects objects.Module
	IP      ip.Module
	Tree    tree.Module
	Events  events.Module
}

type Settings struct {
	Enabled *tree.Value[*astral.Bool] `tree:"enabled"`
}

// Module is the concrete implementation of the NAT module.
type Module struct {
	Deps

	ctx      *astral.Context
	node     astral.Node
	log      *log.Logger
	assets   resources.Resources
	settings Settings

	pool *PairPool
	ops  ops.Set

	enabled atomic.Bool
	cond    *sync.Cond
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)

	go func() {
		for range mod.settings.Enabled.Follow(ctx) {
			mod.evaluateEnabled()
		}
	}()

	<-ctx.Done()

	return nil
}

func (mod *Module) evaluateEnabled() {
	setting := mod.settings.Enabled.Get()
	settingEnabled := setting == nil || bool(*setting)
	hasPublicIPs := len(mod.IP.PublicIPCandidates()) > 0

	mod.SetEnabled(settingEnabled && hasPublicIPs)
}

func (mod *Module) GetOpSet() *ops.Set {
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

func (mod *Module) newPuncher(session []byte) (nat.Puncher, error) {
	cb := &ConePuncherCallbacks{
		OnAttempt:       func(peer ip.IP, port int, _ []*net.UDPAddr) { mod.log.Log("punching → %v:%v", peer, port) },
		OnProbeReceived: func(from *net.UDPAddr) { mod.log.Log("probe ← %v", from) },
	}
	p, err := newConePuncher(session, cb)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (mod *Module) getLocalIPv4() (ip.IP, error) {
	for _, addr := range mod.IP.PublicIPCandidates() {
		if addr.IsIPv4() {
			return addr, nil
		}
	}
	return nil, nat.ErrNoSuitableIP
}
