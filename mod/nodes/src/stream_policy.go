package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

// StreamPolicy is the single abstraction for stream lifecycle management.
// It can limit outbound links, timeout idle links, react to network changes, and evict/rebalance streams.
type StreamPolicy interface {
	// Before dialing any endpoint (prevents useless handshakes)
	OnBeforeDial(remoteID *astral.Identity, ep exonet.Endpoint) Decision
	// After handshake, before admission (may reject or evict)
	OnBeforeAdmit(s *Stream) Decision
	// Called when stream activity occurs (read/write)
	OnActivity(s *Stream, dir string)
	// Optional: network change / gateway updates
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
