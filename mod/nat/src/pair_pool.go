package nat

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/sig"
)

type PairPool struct {
	*Module
	pairs sig.Map[astral.Nonce, *pairEntry]
	stop  chan struct{}
}

var _ nat.PairPool = &PairPool{}

func NewPairPool() *PairPool {
	return &PairPool{stop: make(chan struct{})}
}

func (p *PairPool) Add(pair *nat.EndpointPair, local *astral.Identity, isPinger bool) error {
	if pair.Nonce == 0 {
		return fmt.Errorf("missing nonce")
	}
	if _, ok := p.pairs.Get(pair.Nonce); ok {
		return fmt.Errorf("duplicate nonce")
	}

	e := &pairEntry{EndpointPair: *pair}
	err := e.init(local, isPinger)
	if err != nil {
		return err
	}

	p.pairs.Set(pair.Nonce, e)
	return nil
}

func (p *PairPool) get(nonce astral.Nonce) (*pairEntry, bool) {
	return p.pairs.Get(nonce)
}

// Take returns an idle pair that matches the given peer identity and locks on both sides.
func (p *PairPool) Take(ctx *astral.Context, peer *astral.Identity) (pair *nat.
	EndpointPair, err error) {
	for _, n := range p.pairs.Keys() {
		if pair, ok := p.pairs.Get(n); ok && pair.isIdle() && pair.matchesPeer(peer) {
			locking := pair.beginLock()
			if !locking {
				return nil, fmt.Errorf("cannot begin to lock pair")
			}

			remoteEndpoint, ok := pair.RemoteEndpoint(ctx.Identity())
			if !ok {
				return nil, fmt.Errorf("cannot find remote endpoint")
			}

			args := opPairTakeArgs{
				Pair: pair.Nonce,
			}

			takeQuery := query.New(
				ctx.Identity(),
				remoteEndpoint.Identity,
				nat.MethodPairTake,
				&args,
			)

			peerCh, err := query.RouteChan(
				ctx.IncludeZone(astral.ZoneNetwork),
				p.node,
				takeQuery,
			)
			if err != nil {
				return nil, fmt.Errorf("cannot route to peer: %w", err)
			}

			pairTaker := NewPairTaker(rolePairTaker, peerCh, pair, remoteEndpoint.Identity)

			err = pairTaker.Run(ctx)
			if err != nil {
				return nil, err
			}

			// Pair is now locked on both sides, we can remove it from the pool.
			p.pairs.Delete(n)

			return &pair.EndpointPair, nil
		}
	}
	return nil, fmt.Errorf("no idle pair found")
}

func (p *PairPool) Remove(nonce astral.Nonce) {
	if e, ok := p.pairs.Delete(nonce); ok {
		e.expire()
	}
}

func (p *PairPool) RunCleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, n := range p.pairs.Keys() {
					if e, ok := p.pairs.Get(n); ok && e.isExpired() {
						p.pairs.Delete(n)
					}
				}
			case <-p.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *PairPool) Stop() { close(p.stop) }

func (p *PairPool) Size() int { return p.pairs.Len() }
