package link

import (
	"fmt"
)

type Event interface{}

type EventConnAdded struct {
	localName  string
	remoteName string
	Conn       *Conn
}

func (e EventConnAdded) String() string {
	return fmt.Sprintf("query=%s outbound=%t localID=%s remoteID=%s",
		e.Conn.Query(),
		e.Conn.Outbound(),
		e.localName,
		e.remoteName,
	)
}

type EventConnRemoved struct {
	localName  string
	remoteName string
	Conn       *Conn
}

func (e EventConnRemoved) String() string {
	return fmt.Sprintf("query=%s outbound=%t localID=%s remoteID=%s",
		e.Conn.Query(),
		e.Conn.Outbound(),
		e.localName,
		e.remoteName,
	)
}

type EventLinkEstablished struct {
	Link *Link
}

func (e EventLinkEstablished) String() string {
	return fmt.Sprintf(
		"network=%s localAddr=%s remoteAddr=%s prio=%d remoteID=%s",
		e.Link.Network(),
		e.Link.LocalEndpoint(),
		e.Link.RemoteEndpoint(),
		e.Link.Priority(),
		e.Link.RemoteIdentity().String(),
	)
}

type EventLinkClosed struct {
	Link *Link
}

func (e EventLinkClosed) String() string {
	return fmt.Sprintf(
		"network=%s localAddr=%s remoteAddr=%s prio=%d err=%s remoteID=%s",
		e.Link.Network(),
		e.Link.LocalEndpoint(),
		e.Link.RemoteEndpoint(),
		e.Link.Priority(),
		e.Link.Err(),
		e.Link.RemoteIdentity().String(),
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
		e.Link.LocalEndpoint(),
		e.Link.RemoteEndpoint(),
		e.Link.Priority(),
		e.Link.RemoteIdentity().String(),
	)
}
