package nodes

import (
	"sort"
)

// NetworkPreferencePolicy prefers higher quality networks (TCP > UTP > ToR)
type NetworkPreferencePolicy struct {
	networkPriority map[string]int // lower number = higher priority
}

func NewNetworkPreferencePolicy() *NetworkPreferencePolicy {
	return &NetworkPreferencePolicy{
		networkPriority: map[string]int{
			"tcp": 1,
			"utp": 2,
			"gw":  3,
			"tor": 4,
		},
	}
}

func (p *NetworkPreferencePolicy) Name() string {
	return "network_preference"
}

func (p *NetworkPreferencePolicy) Evaluate(state StreamState) PolicyDecision {
	// Group streams by remote identity
	streamsByIdentity := make(map[string][]*Stream)
	for _, s := range state.AllStreams {
		identity := s.RemoteIdentity().String()
		streamsByIdentity[identity] = append(streamsByIdentity[identity], s)
	}

	var actions []PolicyAction

	// For each identity, close streams on inferior networks if better ones exist
	for _, streams := range streamsByIdentity {
		if len(streams) <= 1 {
			continue // Only one stream, nothing to compare
		}

		// Sort by network priority (best first)
		sort.Slice(streams, func(i, j int) bool {
			return p.getNetworkPriority(streams[i].Network()) < p.getNetworkPriority(streams[j].Network())
		})

		// Keep the best network stream, close the rest
		bestNetwork := streams[0].Network()
		for i := 1; i < len(streams); i++ {
			if streams[i].Network() != bestNetwork {
				actions = append(actions, NewCloseStreamAction(streams[i], p.Name()))
			}
		}
	}

	return PolicyDecision{Actions: actions}
}

func (p *NetworkPreferencePolicy) getNetworkPriority(network string) int {
	if priority, ok := p.networkPriority[network]; ok {
		return priority
	}

	return 999 // unknown networks have lowest priority
}
