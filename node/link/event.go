package link

type Event interface{}

type EventConnEstablished struct {
	Conn *Conn
}

type EventConnClosed struct {
	Conn *Conn
}

type EventPingTimeout struct {
	Link *Link
}
