package nat

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
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

func (p *PairPool) Get(nonce astral.Nonce) *pairEntry {
	e, _ := p.pairs.Get(nonce)
	return e
}

func (p *PairPool) Take(peer *astral.Identity) *nat.EndpointPair {
	for _, n := range p.pairs.Keys() {
		if e, ok := p.pairs.Get(n); ok && e.isIdle() && e.matchesPeer(peer) && e.beginLock() {
			// TODO: start coordination with peer about lock

			return e
		}
	}
	return nil
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
