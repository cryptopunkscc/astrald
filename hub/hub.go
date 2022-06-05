package hub

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

const queryResponseTimeout = 5 * time.Second

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

	//TODO: Emit an event for logging?

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
func (hub *Hub) Query(query string, link *link.Link) (*Conn, error) {
	// Fetch the port
	port, err := hub.getPort(query)
	if err != nil {
		return nil, err
	}

	// pass the query to the port
	q := NewQuery(query, link)
	select {
	case port.queries <- q:
	default:
		return nil, ErrQueueOverflow
	}

	// Wait for the response
	var accepted bool
	select {
	case accepted = <-q.response:
	case <-time.After(queryResponseTimeout):
		return nil, ErrTimeout
	}

	if !accepted {
		return nil, ErrRejected
	}

	// Create a pipe for the caller and the responder
	clientConn, appConn := connPipe(query, link)

	// Send one side to the responder
	q.connection <- &appConn
	close(q.connection)

	// Return the other side to the caller
	return &clientConn, nil
}

// release closes a port in the hub
func (hub *Hub) release(name string) error {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	port, found := hub.ports[name]
	if !found {
		return ErrPortNotFound
	}

	close(port.queries)
	delete(hub.ports, name)

	//TODO: Emit an event for logging?
	//log.Println("port released:", name)

	return nil
}

func (hub *Hub) getPort(name string) (*Port, error) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// Fetch the port
	port, found := hub.ports[name]
	if !found {
		return nil, ErrPortNotFound
	}

	return port, nil
}
