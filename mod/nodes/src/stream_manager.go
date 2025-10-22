package nodes

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

// StreamState provides all necessary inputs for a policy decision.
// StreamManager prepares this state by snapshotting current streams.
// Policies can filter AllStreams internally based on their logic (per-identity, global, etc.)
type StreamState struct {
	Candidate  *Stream   // The newly admitted stream
	AllStreams []*Stream // Snapshot of all active streams (including candidate)
}

// PolicyDecision describes the policy's decision.
type PolicyDecision struct {
	Actions []PolicyAction
}

// Change PolicyAction from a func type to an interface executed against a StreamControl
// to align with struct-based actions implementing Execute.
type PolicyAction interface {
	Execute(ctrl StreamControl) error
}

type StreamControl interface {
	CloseStream(s *Stream) error
	ProtectStream(s *Stream) error
	IsProtected(s *Stream) bool
}

// StreamPolicy defines a policy hook used by StreamManager.
// Policies are pure functions: they operate only on inputs (StreamState) and return outputs (PolicyDecision).
// They do NOT access module state, Peers, or any external dependencies directly.
// Policies can internally filter state.AllStreams to their desired scope (per-identity, global, etc.)
type StreamPolicy interface {
	// Evaluate runs the policy logic against the provided state.
	// It may propose actions to manage streams (close, protect, balance, etc.)
	Evaluate(state StreamState) PolicyDecision
	// Name returns a human-readable policy name for logging.
	Name() string
}

// StreamManager orchestrates post-admission stream management policies.
type StreamManager struct {
	log      *log.Logger
	policies []StreamPolicy
}

// NewStreamManager creates a new StreamManager with configured policies.
func NewStreamManager(log *log.Logger) *StreamManager {
	sm := &StreamManager{
		log: log,
	}

	// Register policies in order
	// Protection policies should come before eviction policies for better conflict resolution
	sm.policies = []StreamPolicy{
		NewSiblingGuardPolicy(),       // Protect minimum sibling connections (default)
		NewActiveStreamGuardPolicy(),  // Protect streams active in recent window (default)
		NewNetworkPreferencePolicy(),  // Prefer better networks
		NewMaxOutboundStreamsPolicy(), // Limit outbound streams (default)
	}

	return sm
}

// Run executes the post-admission policy pipeline.
// It snapshots current streams, evaluates policies, merges proposals, and executes actions.
func (sm *StreamManager) Run(candidate *Stream, allStreams []*Stream) []*Stream {
	// Build state snapshot
	state := StreamState{
		Candidate:  candidate,
		AllStreams: allStreams,
	}

	// Create a fresh controller for this run so protections are per-run
	sc := NewStreamController()

	// Phase 1: Collect all policy proposals
	var allProposals []PolicyDecision
	for _, policy := range sm.policies {
		proposal := policy.Evaluate(state)

		if len(proposal.Actions) > 0 {
			sm.log.Logv(2, "%s proposed %d actions",
				policy.Name(), len(proposal.Actions))
		}

		allProposals = append(allProposals, proposal)
	}

	// Phase 2: Merge proposals with conflict resolution
	mergedActions := sm.mergeProposals(allProposals)

	// Phase 3: Execute merged actions (with reasons)
	var closedStreams []*Stream
	for _, action := range mergedActions {
		switch a := action.(type) {
		case *ProtectStreamAction:
			sm.log.Logv(2, "protecting stream %d to %v (%v) reasons=[%s]",
				a.Stream.id, a.Stream.RemoteIdentity(), a.Stream.Network(), strings.Join(a.Reasons, ","))
		case *CloseStreamAction:
			sm.log.Infov(1, "closing stream %d to %v (%v) reasons=[%s]",
				a.Stream.id, a.Stream.RemoteIdentity(), a.Stream.Network(), strings.Join(a.Reasons, ","))
		}

		if err := action.Execute(sc); err != nil {
			sm.log.Errorv(2, "failed to execute action: %v", err)
		}

		// Track closed streams for return value
		if closeAction, ok := action.(*CloseStreamAction); ok {
			closedStreams = append(closedStreams, closeAction.Stream)
		}
	}

	if len(closedStreams) > 0 {
		sm.log.Logv(1, "closed %d streams due to policies", len(closedStreams))
	}

	return closedStreams
}

// mergeProposals merges all policy proposals using conflict resolution rules:
// Protection > Eviction > None
// If any policy protects a stream, it cannot be evicted.
func (sm *StreamManager) mergeProposals(proposals []PolicyDecision) []PolicyAction {
	protectSet := make(map[int]*ProtectStreamAction) // streamID → protect action (reasons aggregated)
	evictSet := make(map[int]*CloseStreamAction)     // streamID → close action (reasons aggregated)

	// Collect all actions by type and aggregate reasons
	for _, proposal := range proposals {
		for _, action := range proposal.Actions {
			switch a := action.(type) {
			case *ProtectStreamAction:
				if existing, ok := protectSet[a.Stream.id]; ok {
					existing.Reasons = dedupeReasons(append(existing.Reasons, a.Reasons...))
				} else {
					protectSet[a.Stream.id] = &ProtectStreamAction{Stream: a.Stream, Reasons: dedupeReasons(a.Reasons)}
				}
			case *CloseStreamAction:
				if existing, ok := evictSet[a.Stream.id]; ok {
					existing.Reasons = dedupeReasons(append(existing.Reasons, a.Reasons...))
				} else {
					evictSet[a.Stream.id] = &CloseStreamAction{Stream: a.Stream, Reasons: dedupeReasons(a.Reasons)}
				}
			}
		}
	}

	// Apply conflict resolution: Protection overrides eviction
	for id := range protectSet {
		delete(evictSet, id)
	}

	// Build final action list
	var merged []PolicyAction
	for _, action := range protectSet {
		merged = append(merged, action)
	}
	for _, a := range evictSet {
		merged = append(merged, a)
	}
	return merged
}

// deduplicateStreams removes duplicate streams from the list.
func deduplicateStreams(streams []*Stream) []*Stream {
	if len(streams) == 0 {
		return nil
	}

	seen := make(map[*Stream]bool)
	result := make([]*Stream, 0, len(streams))

	for _, s := range streams {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// filterByIdentity is a helper function that filters streams by remote identity.
// Policies can use this internally to limit their scope to a specific identity.
func filterByIdentity(streams []*Stream, identity *astral.Identity) []*Stream {
	var filtered []*Stream
	for _, s := range streams {
		if s.RemoteIdentity().IsEqual(identity) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
