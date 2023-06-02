package link

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"time"
)

func (l *Link) Activity() *sig.Activity {
	return &l.activity
}

func (l *Link) Idle() *Idle {
	return l.idle
}

func (l *Link) SetQueryHandler(queryHandler QueryHandlerFunc) {
	l.queryHandler = queryHandler
}

func (l *Link) Events() *events.Queue {
	return &l.events
}

func (l *Link) Err() error {
	return l.err
}

func (l *Link) EstablishedAt() time.Time {
	return l.establishedAt
}

func (l *Link) Done() <-chan struct{} {
	return l.doneCh
}

func (l *Link) Conns() *ConnSet {
	return l.conns
}

func (l *Link) Priority() int {
	return l.priority
}

func (l *Link) SetPriority(priority int) {
	if l.priority == priority {
		return
	}

	var e = EventLinkPriorityChanged{
		Link: l,
		Old:  l.priority,
		New:  priority,
	}
	defer l.events.Emit(e)

	l.priority = priority
}

func (l *Link) Mux() *mux.FrameMux {
	return l.mux
}

func (l *Link) Health() *Health {
	return l.health
}

func (l *Link) Outbound() bool {
	return l.conn.Outbound()
}

func (l *Link) Network() string {
	if l.conn.LocalEndpoint() != nil {
		return l.conn.LocalEndpoint().Network()
	}
	if l.conn.RemoteEndpoint() != nil {
		return l.conn.RemoteEndpoint().Network()
	}
	return ""
}

func (l *Link) RemoteEndpoint() net.Endpoint {
	return l.conn.RemoteEndpoint()
}

func (l *Link) LocalEndpoint() net.Endpoint {
	return l.conn.LocalEndpoint()
}

func (l *Link) RemoteIdentity() id.Identity {
	return l.conn.RemoteIdentity()
}

func (l *Link) LocalIdentity() id.Identity {
	return l.conn.LocalIdentity()
}
