package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/sig"
)

type HolePool struct {
	*Module
	holes sig.Map[astral.Nonce, *Hole]
	stop  chan struct{}
}

func NewHolePool(mod *Module) *HolePool {
	return &HolePool{
		Module: mod,
		stop:   make(chan struct{}),
		holes:  sig.Map[astral.Nonce, *Hole]{},
	}
}

func (p *HolePool) Add(hole *Hole) error {
	_, ok := p.holes.Set(hole.Nonce, hole)
	if !ok {
		return nat.ErrDuplicateHole
	}
	return nil
}

func (p *HolePool) Get(nonce astral.Nonce) (*Hole, bool) {
	return p.holes.Get(nonce)
}

func (p *HolePool) GetAll() []*Hole {
	var holes []*Hole
	for _, hole := range p.holes.Values() {
		if !hole.IsIdle() {
			continue
		}
		holes = append(holes, hole)
	}
	return holes
}

func (p *HolePool) TakeAny(peer *astral.Identity) (*Hole, error) {
	for _, hole := range p.holes.Values() {
		if hole.MatchesPeer(peer) && hole.IsIdle() {
			p.Remove(hole.Nonce)
			return hole, nil
		}
	}
	return nil, nat.ErrHoleNotExists
}

func (p *HolePool) Take(nonce astral.Nonce) (*Hole, error) {
	hole, ok := p.holes.Get(nonce)
	if !ok {
		return nil, nat.ErrHoleNotExists
	}

	if !hole.IsIdle() {
		return nil, nat.ErrHoleBusy
	}

	p.Remove(nonce)
	return hole, nil
}

func (p *HolePool) Remove(nonce astral.Nonce) (hole *Hole, ok bool) {
	return p.holes.Delete(nonce)
}
