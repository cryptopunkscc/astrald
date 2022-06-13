package hub

type EventPortRegistered struct {
	PortName string
}

type EventPortReleased struct {
	PortName string
}
