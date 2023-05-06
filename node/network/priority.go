package network

import (
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	"github.com/cryptopunkscc/astrald/node/link"
)

var networkPriorities map[string]int

// BestQuality selects the best available link.
func BestQuality(current *link.Link, next *link.Link) *link.Link {
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

	if current.Network() == inet.DriverName {
		currentAddr := current.RemoteEndpoint().(inet.Endpoint)
		nextAddr := next.RemoteEndpoint().(inet.Endpoint)

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
	if next.Health().LastRTT() < current.Health().LastRTT() {
		return next
	}

	return current
}

// NetworkPriority returns network's priority
func NetworkPriority(netName string) int {
	return networkPriorities[netName]
}

func init() {
	networkPriorities = map[string]int{
		"inet": 400,
		"bt":   300,
		"gw":   200,
		"tor":  100,
	}
}
