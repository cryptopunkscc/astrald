package nodes

import (
	"sort"
)

// MaxOutboundStreamsPolicy limits the maximum number of outbound streams
type MaxOutboundStreamsPolicy struct {
	maxOutbound int
}

const defaultMaxOutbound = 10

func NewMaxOutboundStreamsPolicy() *MaxOutboundStreamsPolicy {
	return &MaxOutboundStreamsPolicy{
		maxOutbound: defaultMaxOutbound,
	}
}

func (p *MaxOutboundStreamsPolicy) Name() string {
	return "max_outbound_streams"
}

func (p *MaxOutboundStreamsPolicy) Evaluate(state StreamState) PolicyDecision {
	// Count outbound streams
	var outboundStreams []*Stream
	for _, s := range state.AllStreams {
		if s.outbound {
			outboundStreams = append(outboundStreams, s)
		}
	}

	// If within limit, no action needed
	if len(outboundStreams) <= p.maxOutbound {
		return PolicyDecision{Actions: nil}
	}

	// Close oldest outbound streams to get back to limit
	// Sort by creation time (oldest first)
	sort.Slice(outboundStreams, func(i, j int) bool {
		return outboundStreams[i].createdAt.Before(outboundStreams[j].createdAt)
	})

	excess := len(outboundStreams) - p.maxOutbound
	var actions []PolicyAction
	for i := 0; i < excess; i++ {
		actions = append(actions, NewCloseStreamAction(outboundStreams[i], p.Name()))
	}

	return PolicyDecision{Actions: actions}
}
