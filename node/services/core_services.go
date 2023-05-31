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

func (m *CoreService) Find(name string) (*Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fetch the service
	service, found := m.services[name]
	if !found {
		return nil, ErrServiceNotFound
	}

	return service, nil
}

func (m *CoreService) Query(ctx context.Context, identity id.Identity, query string, link *link.Link) (*Conn, error) {
	// Fetch the service
	service, err := m.Find(query)
	if err != nil {
		return nil, err
	}

	cliConn, srvConn := pipe(query, link)
	srvConn.remoteID = identity
	cliConn.remoteID = service.Identity()

	// pass the query to the service
	var q = newQuery(query, link, identity, srvConn)

	if service.handler == nil {
		panic("service has nil handler")
	}

	service.handler(ctx, q)

	select {
	case accepted := <-q.response:
		if accepted {
			return cliConn, nil
		}
		return nil, ErrRejected

	case <-ctx.Done():
		q.cancel(ctx.Err())
		return nil, ctx.Err()

	case <-time.After(queryResponseTimeout):
		q.cancel(ErrTimeout)
		return nil, ErrTimeout
	}

}

// Register registers a service as the specified identity
func (m *CoreService) Register(ctx context.Context, identity id.Identity, name string, handler QueryHandlerFunc) (*Service, error) {
	service, err := m.register(name, identity, handler)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		service.Close()
	}()

	return service, nil
}

func (m *CoreService) register(name string, identity id.Identity, handler QueryHandlerFunc) (*Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the requested service is available
	if _, found := m.services[name]; found {
		return nil, ErrAlreadyRegistered
	}

	// register the service
	m.services[name] = newService(m, identity, name, handler)

	m.log.Infov(1, "service %s registered", name)

	m.events.Emit(EventServiceRegistered{name})

	return m.services[name], nil
}

// release removes a service from the manager
func (m *CoreService) release(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, found := m.services[name]
	if !found {
		return ErrServiceNotFound
	}

	delete(m.services, name)

	m.log.Infov(1, "service %s released", name)

	m.events.Emit(EventServiceReleased{name})

	return nil
}
