package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"sync"
	"time"
)

const queryResponseTimeout = time.Second
const logTag = "services"

var _ Services = &CoreServices{}
var _ net.Router = &CoreServices{}

// CoreServices facilitates registration of services and querying them.
type CoreServices struct {
	services map[string]*Service
	mu       sync.Mutex
	events   events.Queue
	log      *log.Logger
}

func NewCoreServices(eventParent *events.Queue, log *log.Logger) *CoreServices {
	hub := &CoreServices{
		services: make(map[string]*Service),
		log:      log.Tag(logTag),
	}
	hub.events.SetParent(eventParent)
	return hub
}

// List returns information about all registered services
func (srv *CoreServices) List() []ServiceInfo {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var list = make([]ServiceInfo, 0, len(srv.services))
	for _, service := range srv.services {
		list = append(list, ServiceInfo{
			Name:         service.name,
			Identity:     service.identity,
			RegisteredAt: service.registeredAt,
		})
	}
	return list
}

func (srv *CoreServices) Find(name string) (*Service, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Fetch the service
	service, found := srv.services[name]
	if !found {
		return nil, ErrServiceNotFound
	}

	return service, nil
}

// Register registers a service as the specified identity
func (srv *CoreServices) Register(ctx context.Context, identity id.Identity, name string, handler net.Router) (*Service, error) {
	service, err := srv.register(name, identity, handler)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		service.Close()
	}()

	return service, nil
}

func (srv *CoreServices) register(name string, identity id.Identity, handler net.Router) (*Service, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	// Check if the requested service is available
	if _, found := srv.services[name]; found {
		return nil, ErrAlreadyRegistered
	}

	// register the service
	srv.services[name] = newService(srv, identity, name, handler)

	srv.log.Infov(1, "registered %v:%s", identity, name)

	srv.events.Emit(EventServiceRegistered{
		Identity: identity,
		Name:     name,
	})

	return srv.services[name], nil
}

// release removes a service from the manager
func (srv *CoreServices) release(name string) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	s, found := srv.services[name]
	if !found {
		return ErrServiceNotFound
	}

	delete(srv.services, name)

	srv.log.Infov(1, "released %v:%s", s.identity, s.name)

	srv.events.Emit(EventServiceReleased{
		Identity: s.identity,
		Name:     s.name,
	})

	return nil
}
