package hub

// Port represents an open port in the hub
type Port struct {
	name     string
	requests chan *Request
	hub      *Hub
}

// Requests returns a channel for reading incoming connection requests
func (port *Port) Requests() <-chan *Request {
	return port.requests
}

// Close closees the port
func (port *Port) Close() error {
	return port.hub.close(port.name)
}
