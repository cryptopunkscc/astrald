package link

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/log"
)

type Event interface{}

type EventConnEstablished struct {
	Conn *Conn
}

func (e EventConnEstablished) String() string {
	return fmt.Sprintf("query=%s localPort=%d outbound=%t", e.Conn.Query(), e.Conn.LocalPort(), e.Conn.Outbound())
}

type EventConnClosed struct {
	Conn *Conn
}

func (e EventConnClosed) String() string {
	return fmt.Sprintf("query=%s localPort=%d outbound=%t err=%s", e.Conn.Query(), e.Conn.LocalPort(), e.Conn.Outbound(), e.Conn.err)
}

type EventLinkEstablished struct {
	Link *Link
}

func (e EventLinkEstablished) String() string {
	return fmt.Sprintf(
		"network=%s localAddr=%s remoteAddr=%s prio=%d remoteID=%s",
		e.Link.Network(),
		e.Link.LocalAddr(),
		e.Link.RemoteAddr(),
		e.Link.Priority(),
		log.Sprint("%s", e.Link.RemoteIdentity()),
	)
}

type EventLinkClosed struct {
	Link *Link
}

func (e EventLinkClosed) String() string {
	return fmt.Sprintf(
		"network=%s localAddr=%s remoteAddr=%s prio=%d err=%s remoteID=%s",
		e.Link.Network(),
		e.Link.LocalAddr(),
		e.Link.RemoteAddr(),
		e.Link.Priority(),
		e.Link.Err(),
		log.Sprint("%s", e.Link.RemoteIdentity()),
	)
}

type EventLinkPriorityChanged struct {
	Link *Link
	Old  int
	New  int
}

func (e EventLinkPriorityChanged) String() string {
	return fmt.Sprintf(
		"network=%s localAddr=%s remoteAddr=%s prio=%d remoteID=%s",
		e.Link.Network(),
		e.Link.LocalAddr(),
		e.Link.RemoteAddr(),
		e.Link.Priority(),
		log.Sprint("%s", e.Link.RemoteIdentity()),
	)
}
