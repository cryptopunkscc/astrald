package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

const queryResponseTimeout = time.Second
const logTag = "services"

var _ Services = &CoreService{}

// CoreService facilitates registration of services and querying them.
type CoreService struct {
	services map[string]*Service
	mu       sync.Mutex
	events   event.Queue
	log      *log.Logger
}

func NewCoreServices(eventParent *event.Queue, log *log.Logger) *CoreService {
	hub := &CoreService{
		services: make(map[string]*Service),
		log:      log.Tag(logTag),
	}
	hub.events.SetParent(eventParent)
	return hub
}

// List returns information about all registered services
func (m *CoreService) List() []ServiceInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	var list = make([]ServiceInfo, 0, len(m.services))
	for _, service := range m.services {
		list = append(list, ServiceInfo{
			Name:         service.name,
			Identity:     service.identity,
			RegisteredAt: service.registeredAt,
		})
	}
	return list
}

// Register registers a service as the default identity
func (m *CoreService) Register(ctx context.Context, name string) (*Service, error) {
	return m.RegisterAs(ctx, name, id.Identity{}) //TODO: this should be node's identity by default
}

// RegisterAs registers a service as the specified identity
func (m *CoreService) RegisterAs(ctx context.Context, name string, identity id.Identity) (*Service, error) {
	service, err := m.register(name, identity)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		service.Close()
	}()

	return service, nil
}

func (m *CoreService) register(name string, identity id.Identity) (*Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the requested service is available
	if _, found := m.services[name]; found {
		return nil, ErrAlreadyRegistered
	}

	// register the service
	m.services[name] = NewService(m, name, identity)

	m.log.Infov(1, "service %s registered", name)

	m.events.Emit(EventServiceRegistered{name})

	return m.services[name], nil
}

func (m *CoreService) Query(ctx context.Context, query string, link *link.Link) (*Conn, error) {
	return m.QueryAs(ctx, query, link, link.RemoteIdentity())
}

func (m *CoreService) QueryAs(ctx context.Context, query string, link *link.Link, identity id.Identity) (*Conn, error) {
	// Fetch the service
	service, err := m.getService(query)
	if err != nil {
		return nil, err
	}

	// pass the query to the service
	q := NewQuery(query, link, identity)
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

	appConn.remoteID = identity

	// Send one side to the responder
	q.connection <- &appConn
	close(q.connection)

	// Return the other side to the caller
	return &clientConn, nil
}

// Release closes a service in the manager
func (m *CoreService) Release(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, found := m.services[name]
	if !found {
		return ErrServiceNotFound
	}

	close(service.queries)
	delete(m.services, name)

	m.log.Infov(1, "service %s released", name)

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
