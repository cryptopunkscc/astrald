package network

import (
	"github.com/cryptopunkscc/astrald/net"
)

var networkPriorities map[string]int

// BestQuality selects the best available link.
func BestQuality(current net.Link, next net.Link) net.Link {
	if current == nil {
		return next
	}

	cscore := getNetworkPriority(net.Network(current))
	nscore := getNetworkPriority(net.Network(next))

	if nscore > cscore {
		return next
	}

	return current
}

// getNetworkPriority returns network's priority
func getNetworkPriority(netName string) int {
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
