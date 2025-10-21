package nodes

// StreamPolicy is the single abstraction for stream lifecycle management.
// It can limit outbound links, timeout idle links, react to network changes, and evict/rebalance streams.
type StreamPolicy interface {
	// OnAdmitted is called immediately after a new stream has been admitted
	// into the Peers manager. Policies may choose to evict redundant or
	// less preferred streams at this point. This method must be non-blocking.
	OnAdmitted(s *Stream) Decision
	// OnActivity is called whenever activity occurs on a stream
	// (read, write, or ping). Used by policies such as IdleTimeoutPolicy
	// to update last activity timestamps.
	OnActivity(s *Stream, dir string)
	// OnRebalance is called when external network conditions change
	// (e.g. gateway updates, interface change). Policies may use this to
	// reevaluate and rebalance streams across networks.
	OnRebalance()
}

// DecisionAction represents the action to take in a policy decision.
type DecisionAction string

const (
	DecisionAllow  DecisionAction = "allow"
	DecisionSkip   DecisionAction = "skip"
	DecisionReject DecisionAction = "reject"
	DecisionEvict  DecisionAction = "evict"
)

// Decision describes the result of a policy hook.
type Decision struct {
	Action  DecisionAction // DecisionAllow | DecisionSkip | DecisionReject | DecisionEvict
	Reason  string
	Victims []*Stream
}
