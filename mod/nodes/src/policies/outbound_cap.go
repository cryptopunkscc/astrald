// filepath: /Users/grzegorz/Documents/projects/work2/astrald/mod/nodes/src
package policies

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	nodes "github.com/cryptopunkscc/astrald/mod/nodes/src"
)

// OutboundCapPolicy is a no-op placeholder. It satisfies nodes.StreamPolicy but performs no actions.
type OutboundCapPolicy struct{}

var _ nodes.StreamPolicy = (*OutboundCapPolicy)(nil)

func (p *OutboundCapPolicy) OnBeforeDial(remoteID *astral.Identity, ep exonet.Endpoint) nodes.Decision {
	// no-op: allow dial
	return nodes.Decision{}
}

func (p *OutboundCapPolicy) OnBeforeAdmit(s *nodes.Stream) nodes.Decision {
	// no-op: allow admit
	return nodes.Decision{}
}

func (p *OutboundCapPolicy) OnActivity(s *nodes.Stream, dir string) {
	// no-op
}

func (p *OutboundCapPolicy) OnRebalance() {
	// no-op
}

///policies/outbound_cap.go
