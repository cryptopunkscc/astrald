package nodes

import (
	"time"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

const ipCacheSize = 8

type ObservedEndpoint struct {
	Endpoint exonet.Endpoint
	IP       ip.IP // NOTE: to not parse from endpoint
	Observed int64
}

func (mod *Module) FindIPCandidates() (list []ip.IP) {
	cache := mod.observedEndpoints.Clone()
	list = make([]ip.IP, 0, len(cache))
	for _, entry := range cache {
		list = append(list, entry.IP)
	}
	return list
}

// AddObservedEndpoint adds or updates an IP in the cache, evicting the oldest if over size.
func (mod *Module) AddObservedEndpoint(endpoint exonet.Endpoint, ip ip.IP) {
	key := endpoint.Address() // TODO: shouldnt be this with protocol (?)
	mod.observedEndpoints.Set(key, ObservedEndpoint{
		Endpoint: endpoint,
		IP:       ip,
		Observed: time.Now().Unix(),
	})

	cache := mod.observedEndpoints.Clone()
	if len(cache) > ipCacheSize {
		// Find the oldest entry
		var oldestKey string
		var oldestTime int64
		first := true
		for k, v := range cache {
			if first || v.Observed < oldestTime {
				oldestTime = v.Observed
				oldestKey = k
				first = false
			}
		}
		if oldestKey != "" {
			mod.observedEndpoints.Delete(oldestKey)
		}
	}
}
