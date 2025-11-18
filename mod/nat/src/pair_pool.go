package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/sig"
)

type PairPool struct {
	*Module
	pairs sig.Map[astral.Nonce, *Pair]
	stop  chan struct{}
}

func NewPairPool(mod *Module) *PairPool {
	return &PairPool{
		Module: mod,
		stop:   make(chan struct{}),
	}
}

func (p *PairPool) Add(pair *Pair) error {
	_, ok := p.pairs.Set(pair.Nonce, pair)
	if !ok {
		return nat.ErrDuplicatePair
	}

	return nil
}

func (p *PairPool) Get(nonce astral.Nonce) (*Pair, bool) {
	return p.pairs.Get(nonce)
}

func (p *PairPool) GetAll() []*Pair {
	var pairs []*Pair
	for _, pair := range p.pairs.Values() {
		if !pair.IsIdle() {
			continue
		}

		pairs = append(pairs, pair)
	}
	return pairs
}

func (p *PairPool) TakeAny(peer *astral.Identity) (*Pair, error) {
	for _, pair := range p.pairs.Values() {
		if pair.MatchesPeer(peer) && pair.IsIdle() {
			p.pairs.Delete(pair.Nonce)
			return pair, nil
		}
	}

	return nil, nat.ErrPairNotExists
}

func (p *PairPool) Take(nonce astral.Nonce) (*Pair, error) {
	pair, ok := p.pairs.Get(nonce)
	if !ok {
		return nil, nat.ErrPairNotExists
	}

	if !pair.IsIdle() {
		return nil, nat.ErrPairBusy
	}

	p.Remove(nonce)
	return pair, nil
}

func (p *PairPool) Remove(nonce astral.Nonce) {
	fmt.Println("remove pair?")
	if e, ok := p.pairs.Delete(nonce); ok {
		e.Expire()
	}
}

func (p *PairPool) Stop() { close(p.stop) }

func (p *PairPool) Size() int { return p.pairs.Len() }
