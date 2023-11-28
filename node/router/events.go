package router

type EventConnAdded struct {
	Conn *MonitoredConn
}

type EventConnRemoved struct {
	Conn *MonitoredConn
}
