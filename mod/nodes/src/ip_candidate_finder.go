package nodes

import (
	"time"

	"github.com/cryptopunkscc/astrald/mod/ip"
)

const ipCacheSize = 8

type ObservedIP struct {
	IP       ip.IP
	Observed int64
}

func (mod *Module) FindIPCandidates() (list []ip.IP) {
	cache := mod.observedIPs.Clone()
	list = make([]ip.IP, 0, len(cache))
	for _, entry := range cache {
		list = append(list, entry.IP)
	}
	return list
}

// AddObservedIP adds or updates an IP in the cache, evicting the oldest if over size.
func (mod *Module) AddObservedIP(candidate ip.IP) {
	key := candidate.String()
	mod.observedIPs.Set(key, ObservedIP{
		IP:       candidate,
		Observed: time.Now().Unix(),
	})

	cache := mod.observedIPs.Clone()
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
			mod.observedIPs.Delete(oldestKey)
		}
	}
}
