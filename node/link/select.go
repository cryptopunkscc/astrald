package link

import (
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

type SelectFunc func(current *Link, next *Link) *Link

func Select(ch <-chan *Link, selectFunc SelectFunc) (selected *Link) {
	for next := range ch {
		if selected == nil {
			selected = next
			continue
		}
		selected = selectFunc(selected, next)
	}
	return
}

func LowestRoundTrip(current *Link, next *Link) *Link {
	if next.Ping() < current.Ping() {
		return next
	}

	return current
}

func MostRecent(current *Link, next *Link) *Link {
	if next.Idle() < current.Idle() {
		return next
	}

	return current
}

func BestQuality(current *Link, next *Link) *Link {
	if netPrio(current) > netPrio(next) {
		return current
	}

	if netPrio(next) > netPrio(current) {
		return next
	}

	if current.Network() == inet.NetworkName {
		currentAddr := current.RemoteAddr().(inet.Addr)
		nextAddr := next.RemoteAddr().(inet.Addr)

		// if one link is in LAN and the other in WAN, prefer the LAN one
		if currentAddr.IsPrivate() != nextAddr.IsPrivate() {
			if currentAddr.IsPrivate() {
				return current
			} else {
				return next
			}
		}

		// if both links are in the same area, prefer the older one
		if current.EstablishedAt().Before(next.EstablishedAt()) {
			return current
		}
		return next
	}

	// for gw and tor just pick best latency
	if next.Ping() < current.Ping() {
		return next
	}

	return current
}

// netPrio returns network priority of a link
func netPrio(l *Link) int {
	switch l.Network() {
	case inet.NetworkName:
		return 40
	case bt.NetworkName:
		return 30
	case gw.NetworkName:
		return 20
	case tor.NetworkName:
		return 10
	}

	return 0
}
