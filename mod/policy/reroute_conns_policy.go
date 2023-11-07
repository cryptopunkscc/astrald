package policy

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
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
		case node.EventConnAdded:
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

func (policy *RerouteConnsPolicy) rerouteConn(ctx context.Context, conn *node.MonitoredConn) {
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

		if (bestScore > currentScore) && (currentLink != best.Link) {
			policy.log.Logv(2, "[%v] rerouting from %v to %v (link %v)",
				nonce,
				currentNet,
				bestNet,
				best.ID(),
			)
			err := policy.modRouter.Reroute(nonce, best.Link)
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
			event, ok := event.(network.EventLinkAdded)
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

func (policy *RerouteConnsPolicy) isReroutable(conn *node.MonitoredConn) bool {
	if _, ok := net.FinalOutput(conn.Target()).(*link.PortWriter); ok {
		return true
	}
	return false
}

func (policy *RerouteConnsPolicy) getLink(conn *node.MonitoredConn) net.Link {
	type linkTest interface {
		Link() *link.CoreLink
	}

	final := net.FinalOutput(conn.Target())
	l, ok := final.(linkTest)
	if !ok {
		return nil
	}
	return l.Link()
}

func (policy *RerouteConnsPolicy) getConnNetwork(conn *node.MonitoredConn) string {
	if t, ok := net.FinalOutput(conn.Target()).(net.Transporter); ok {
		return policy.getTransportNetwork(t.Transport())
	}
	return ""
}

func (policy *RerouteConnsPolicy) getTransportNetwork(t net.SecureConn) string {
	if t == nil {
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
