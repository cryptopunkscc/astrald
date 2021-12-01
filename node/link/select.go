package link

import (
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

type SelectFunc func(current *Link, next *Link) *Link

func Select(ch <-chan *Link, selectFunc SelectFunc) (selected *Link) {
	for next := range ch {
		selected = selectFunc(selected, next)
	}
	return
}

func Fastest(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if current.Network() == tor.NetworkName {
		return next
	}

	if current.Network() == iastral.NetworkName {
		if next.Network() == inet.NetworkName {
			return next
		}
	}

	return current
}

func MostRecent(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Idle() < current.Idle() {
		return next
	}

	return current
}
