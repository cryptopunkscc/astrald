package nodes

// SiblingGuardPolicy protects a minimum number of streams to sibling nodes
type SiblingGuardPolicy struct {
	minSiblings int
}

// defaultMinSiblings is the built-in threshold used by the policy
const defaultMinSiblings = 3

func NewSiblingGuardPolicy() *SiblingGuardPolicy {
	return &SiblingGuardPolicy{
		minSiblings: defaultMinSiblings,
	}
}

func (p *SiblingGuardPolicy) Name() string {
	return "sibling_guard"
}

func (p *SiblingGuardPolicy) Evaluate(state StreamState) PolicyDecision {
	// Group streams by remote identity (a "link" is at least one stream per identity)
	streamsByIdentity := make(map[string][]*Stream)
	for _, s := range state.AllStreams {
		id := s.RemoteIdentity().String()
		streamsByIdentity[id] = append(streamsByIdentity[id], s)
	}

	uniqueLinks := len(streamsByIdentity)
	if uniqueLinks == 0 {
		return PolicyDecision{Actions: nil}
	}

	// If the number of unique links is at or below the minimum, protect one stream per identity
	if uniqueLinks <= p.minSiblings {
		var actions []PolicyAction
		for _, group := range streamsByIdentity {
			best := mostRecentlyActive(group)
			if best != nil {
				actions = append(actions, NewProtectStreamAction(best, p.Name()))
			}
		}
		return PolicyDecision{Actions: actions}
	}

	return PolicyDecision{Actions: nil}
}

// mostRecentlyActive returns the stream with the most recent lastActivity from the slice
func mostRecentlyActive(streams []*Stream) *Stream {
	if len(streams) == 0 {
		return nil
	}
	best := streams[0]
	for i := 1; i < len(streams); i++ {
		if streams[i].lastActivity.After(best.lastActivity) {
			best = streams[i]
		}
	}
	return best
}
