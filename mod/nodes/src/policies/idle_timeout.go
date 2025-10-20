// filepath: /Users/grzegorz/Documents/projects/work2/astrald/mod/nodes/src/policies/idle_timeout.go
package policies

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	nodes "github.com/cryptopunkscc/astrald/mod/nodes/src"
)

// IdleTimeoutPolicy is a no-op placeholder. It satisfies nodes.StreamPolicy but performs no actions.
type IdleTimeoutPolicy struct{}

var _ nodes.StreamPolicy = (*IdleTimeoutPolicy)(nil)

func (p *IdleTimeoutPolicy) OnBeforeDial(remoteID *astral.Identity, ep exonet.Endpoint) nodes.Decision {
	// no-op: allow dial
	return nodes.Decision{}
}

func (p *IdleTimeoutPolicy) OnBeforeAdmit(s *nodes.Stream) nodes.Decision {
	// no-op: allow admit
	return nodes.Decision{}
}

func (p *IdleTimeoutPolicy) OnActivity(s *nodes.Stream, dir string) {
	// no-op
}

func (p *IdleTimeoutPolicy) OnRebalance() {
	// no-op
}
