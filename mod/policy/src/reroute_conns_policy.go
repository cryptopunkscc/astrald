package policy

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

// RerouteConnsPolicy reroutes connections to the best available link
type RerouteConnsPolicy struct {
	*Module
}

func NewRerouteConnsPolicy(mod *Module) *RerouteConnsPolicy {
	return &RerouteConnsPolicy{Module: mod}
}

func (policy *RerouteConnsPolicy) Run(ctx context.Context) error {
	for event := range policy.node.Events().Subscribe(ctx) {
		switch event := event.(type) {
		case core.EventConnAdded:
			if policy.isReroutable(event.Conn) {
				go policy.rerouteConn(ctx, event.Conn)
			}
		}
	}
	return nil
}

func (policy *RerouteConnsPolicy) Name() string {
	return "reroute_conns"
}

func (policy *RerouteConnsPolicy) rerouteConn(ctx context.Context, conn *core.MonitoredConn) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// don't optimize the connection for the first 3 seconds to skip short request-response queries
	select {
	case <-ctx.Done():
		return
	case <-conn.Done():
		return
	case <-time.After(3 * time.Second):
	}

	var remoteID = net.FinalOutput(conn.Target()).Identity()
	var nonce = conn.Query().Nonce()
	var trigger = policy.watchLinksWith(ctx, remoteID)

	for {
		var currentNet = policy.getConnNetwork(conn)
		var currentScore = scoreNetwork(currentNet)
		var currentLink = policy.getLink(conn)
		var links = policy.node.Network().Links().ByRemoteIdentity(remoteID).All()
		if len(links) == 0 { // peer unlinked
			return
		}
		var best, bestScore = bestLinkScore(links)
		var bestNet = policy.getTransportNetwork(best.Transport())

		if (bestScore > currentScore) && (currentLink != best) {
			policy.log.Logv(2, "[%v] rerouting from %v to %v (link %v)",
				nonce,
				currentNet,
				bestNet,
				best.ID(),
			)
			err := policy.relay.Reroute(nonce, best)
			if err != nil {
				policy.log.Errorv(1, "[%v] error rerouting: %v", nonce, err)
			} else {
				policy.log.Infov(1, "[%v] rerouted from %s to %s (link %v)",
					nonce,
					currentNet,
					bestNet,
					best.ID(),
				)
			}
		}

		select {
		case <-trigger:
			continue
		case <-ctx.Done():
			return
		case <-conn.Done():
			return
		}
	}
}

func (policy *RerouteConnsPolicy) watchLinksWith(ctx context.Context, identity id.Identity) <-chan struct{} {
	var ch = make(chan struct{}, 0)

	go func() {
		defer close(ch)
		for event := range policy.node.Events().Subscribe(ctx) {
			event, ok := event.(core.EventLinkAdded)
			if !ok {
				continue
			}

			if event.Link.RemoteIdentity().IsEqual(identity) {
				select {
				case <-ctx.Done():
					return
				case ch <- struct{}{}:
				}
			}
		}
	}()

	return ch
}

func (policy *RerouteConnsPolicy) isReroutable(conn *core.MonitoredConn) bool {
	if _, ok := net.FinalOutput(conn.Target()).(*muxlink.PortWriter); ok {
		return true
	}
	return false
}

func (policy *RerouteConnsPolicy) getLink(conn *core.MonitoredConn) net.Link {
	type linkTest interface {
		Link() *muxlink.Link
	}

	final := net.FinalOutput(conn.Target())
	l, ok := final.(linkTest)
	if !ok {
		return nil
	}
	return l.Link()
}

func (policy *RerouteConnsPolicy) getConnNetwork(conn *core.MonitoredConn) string {
	if t, ok := net.FinalOutput(conn.Target()).(exonet.Transporter); ok {
		return policy.getTransportNetwork(t.Transport())
	}
	return ""
}

func (policy *RerouteConnsPolicy) getTransportNetwork(c any) string {
	if c == nil {
		return ""
	}

	t, ok := c.(exonet.Conn)
	if !ok {
		return ""
	}

	if t.LocalEndpoint() != nil {
		return t.LocalEndpoint().Network()
	}
	if t.RemoteEndpoint() != nil {
		return t.RemoteEndpoint().Network()
	}
	return ""
}
