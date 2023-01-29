package link

import (
	"github.com/cryptopunkscc/astrald/infra/inet"
)

type SelectFunc func(current *Link, next *Link) *Link

// Select selects a link from the link array using the provided select function.
func Select(links []*Link, selectFunc SelectFunc) (selected *Link) {
	for _, next := range links {
		selected = selectFunc(selected, next)
	}
	return
}

// LowestPing selects the link with the lowest ping
func LowestPing(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Ping() < current.Ping() {
		return next
	}

	return current
}

// MostRecent selects the link with the shortest idle duration
func MostRecent(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Idle() < current.Idle() {
		return next
	}

	return current
}

// BestQuality selects the best available link.
func BestQuality(current *Link, next *Link) *Link {
	if current == nil {
		if next.Priority() < 0 {
			return nil
		}
		return next
	}

	if current.Priority() > next.Priority() {
		return current
	}

	if next.Priority() > current.Priority() {
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

// HighestPriority selects the link with the highest priority
func HighestPriority(current *Link, next *Link) *Link {
	if current == nil {
		return next
	}

	if next.Priority() > current.Priority() {
		return next
	}

	return current
}
