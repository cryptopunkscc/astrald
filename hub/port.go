package hub

const queryQueueSize = 4

// Port represents an open port in the hub
type Port struct {
	name    string
	queries chan *Query
	hub     *Hub
}

func NewPort(hub *Hub, name string) *Port {
	return &Port{
		name:    name,
		hub:     hub,
		queries: make(chan *Query, queryQueueSize),
	}
}

// Queries returns a channel for reading incoming queries
func (port *Port) Queries() <-chan *Query {
	return port.queries
}

// Close closees the port
func (port *Port) Close() error {
	return port.hub.release(port.name)
}

// Name returns port's name
func (port *Port) Name() string {
	return port.name
}
