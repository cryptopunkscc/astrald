package hub

import (
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"io"
	"log"
	"sync"
)

// Hub facilitates registration of ports and making connections to them.
type Hub struct {
	ports map[string]Port
	mu    sync.Mutex
}

// Register reserves a port with the requested name and returns its handler.
func (hub *Hub) Register(name string) (Port, error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Initiate port map if necessary
	if hub.ports == nil {
		hub.ports = make(map[string]Port)
	}

	// Check if the requested port is free
	if _, found := hub.ports[name]; found {
		return Port{}, ErrAlreadyRegistered
	}

	// Register the port
	hub.ports[name] = Port{
		name:     name,
		requests: make(chan *Request),
		hub:      hub,
	}

	log.Println("port registered:", name)

	return hub.ports[name], nil
}

// Connect requests to connect to a port as the provided auth.Identity
func (hub *Hub) Connect(name string, caller id.Identity) (io.ReadWriteCloser, error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Fetch the port
	port, found := hub.ports[name]
	if !found {
		return nil, ErrPortNotFound
	}

	// Send the request
	request := &Request{
		caller:     caller,
		response:   make(chan bool, 1),
		connection: make(chan Conn, 1),
		query:      name,
	}
	port.requests <- request

	// Wait for the response
	accepted := <-request.response
	if !accepted {
		return nil, ErrRejected
	}

	// Create a pipe for the caller and the responder
	reqConn, resConn := pipe()

	// Send one side to the responder
	request.connection <- resConn
	close(request.connection)

	// Return the other side to the caller
	return reqConn, nil
}

// close closes a port in the hub
func (hub *Hub) close(name string) error {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	port, found := hub.ports[name]
	if !found {
		return ErrPortNotFound
	}

	close(port.requests)
	delete(hub.ports, name)

	log.Println("port released:", name)

	return nil
}
