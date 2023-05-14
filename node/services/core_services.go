package services

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

var log = _log.Tag("services")

const queryResponseTimeout = time.Second

// CoreService facilitates registration of services and querying them.
type CoreService struct {
	services map[string]*Service
	mu       sync.Mutex
	events   event.Queue
}

func NewCoreServices(eventParent *event.Queue) *CoreService {
	hub := &CoreService{
		services: make(map[string]*Service),
	}
	hub.events.SetParent(eventParent)
	return hub
}

// Register registers a service with the provided name and returns its handler.
func (m *CoreService) Register(name string) (*Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the requested service is available
	if _, found := m.services[name]; found {
		return nil, ErrAlreadyRegistered
	}

	// Register the service
	m.services[name] = NewService(m, name)

	log.Infov(1, "service %s registered", name)

	m.events.Emit(EventServiceRegistered{name})

	return m.services[name], nil
}

func (m *CoreService) RegisterContext(ctx context.Context, name string) (*Service, error) {
	service, err := m.Register(name)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		service.Close()
	}()

	return service, nil
}

// Query sends a query to a service as the provided auth.Identity
func (m *CoreService) Query(ctx context.Context, query string, link *link.Link) (*Conn, error) {
	// Fetch the service
	service, err := m.getService(query)
	if err != nil {
		return nil, err
	}

	// pass the query to the service
	q := NewQuery(query, link)
	select {
	case service.queries <- q:

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

// release closes a service in the manager
func (m *CoreService) release(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, found := m.services[name]
	if !found {
		return ErrServiceNotFound
	}

	close(service.queries)
	delete(m.services, name)

	log.Infov(1, "service %s released", name)

	m.events.Emit(EventServiceReleased{name})

	return nil
}

func (m *CoreService) getService(name string) (*Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fetch the service
	service, found := m.services[name]
	if !found {
		return nil, ErrServiceNotFound
	}

	return service, nil
}
