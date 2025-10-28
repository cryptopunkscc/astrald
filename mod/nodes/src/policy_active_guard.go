package nodes

import "time"

// ActiveStreamGuardPolicy protects streams that have recent traffic
type ActiveStreamGuardPolicy struct {
	activityWindow time.Duration
}

const defaultActivityWindow = 5 * time.Minute

func NewActiveStreamGuardPolicy() *ActiveStreamGuardPolicy {
	return &ActiveStreamGuardPolicy{
		activityWindow: defaultActivityWindow,
	}
}

func (p *ActiveStreamGuardPolicy) Name() string {
	return "active_stream_guard"
}

func (p *ActiveStreamGuardPolicy) Evaluate(state StreamState) PolicyDecision {
	var actions []PolicyAction

	// Protect all streams with recent activity
	for _, s := range state.AllStreams {
		if time.Since(s.lastActivity) < p.activityWindow {
			actions = append(actions, NewProtectStreamAction(s, p.Name()))
		}
	}

	return PolicyDecision{Actions: actions}
}
