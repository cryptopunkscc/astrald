package hub

import (
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"io"
	"log"
	"sync"
)

// Hub facilitates registration of ports and making connections to them.
type Hub struct {
	ports map[string]*Port
	mu    sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		ports: make(map[string]*Port),
	}
}

// Register reserves a port with the requested name and returns its handler.
func (hub *Hub) Register(name string) (*Port, error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Check if the requested port is free
	if _, found := hub.ports[name]; found {
		return nil, ErrAlreadyRegistered
	}

	// Register the port
	hub.ports[name] = NewPort(hub, name)

	log.Println("port open:", name)

	return hub.ports[name], nil
}

// Connect requests to connect to a port as the provided auth.Identity
func (hub *Hub) Connect(query string, caller id.Identity) (io.ReadWriteCloser, error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Fetch the port
	port, found := hub.ports[query]
	if !found {
		return nil, ErrPortNotFound
	}

	// Send the request
	request := NewRequest(caller, query)
	port.requests <- request

	// Wait for the response
	accepted := <-request.response
	if !accepted {
		return nil, ErrRejected
	}

	// Create a pipe for the caller and the responder
	clientConn, appConn := pipe()

	// Send one side to the responder
	request.connection <- appConn
	close(request.connection)

	// Return the other side to the caller
	return clientConn, nil
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
