package hub

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

const queryResponseTimeout = time.Second

// Hub facilitates registration of ports and making connections to them.
type Hub struct {
	ports  map[string]*Port
	mu     sync.Mutex
	events event.Queue
}

func New(eventParent *event.Queue) *Hub {
	hub := &Hub{
		ports: make(map[string]*Port),
	}
	hub.events.SetParent(eventParent)
	return hub
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

	hub.events.Emit(EventPortRegistered{name})

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
func (hub *Hub) Query(ctx context.Context, query string, link *link.Link) (*Conn, error) {
	// Fetch the port
	port, err := hub.getPort(query)
	if err != nil {
		return nil, err
	}

	// pass the query to the port
	q := NewQuery(query, link)
	select {
	case port.queries <- q:

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		return nil, ErrQueueOverflow
	}

	// Wait for the response
	var accepted bool
	select {
	case accepted = <-q.response:

	case <-ctx.Done():
		q.setError(ctx.Err())
		return nil, ctx.Err()

	case <-time.After(queryResponseTimeout):
		q.setError(ErrTimeout)
		return nil, ErrTimeout
	}

	if !accepted {
		return nil, ErrRejected
	}

	// Create a pipe for the caller and the responder
	clientConn, appConn := pipe(query, link)

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

	hub.events.Emit(EventPortReleased{name})

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
