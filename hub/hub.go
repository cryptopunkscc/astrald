package hub

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
	"log"
	"sync"
)

// Hub facilitates registration of ports and making connections to them.
type Hub struct {
	ports map[string]*Port
	mu    sync.Mutex
}

func New() *Hub {
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

	return hub.ports[name], nil
}

func (hub *Hub) RegisterContext(ctx context.Context, name string) (*Port, error) {
	port, err := hub.Register(name)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		port.Close()
	}()

	return port, nil
}

// Query requests to connect to a port as the provided auth.Identity
func (hub *Hub) Query(queryString string, caller id.Identity) (io.ReadWriteCloser, error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Fetch the port
	port, found := hub.ports[queryString]
	if !found {
		return nil, ErrPortNotFound
	}

	// Send the request
	query := NewQuery(caller, queryString)
	port.queries <- query

	// Wait for the response
	accepted := <-query.response
	if !accepted {
		return nil, ErrRejected
	}

	// Create a pipe for the caller and the responder
	clientConn, appConn := pipe()

	// Send one side to the responder
	query.connection <- appConn
	close(query.connection)

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

	close(port.queries)
	delete(hub.ports, name)

	log.Println("port released:", name)

	return nil
}
