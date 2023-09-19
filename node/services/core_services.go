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
	services []*Service
	mu       sync.Mutex
	events   events.Queue
	log      *log.Logger
}

func NewCoreServices(eventParent *events.Queue, log *log.Logger) *CoreServices {
	hub := &CoreServices{
		services: make([]*Service, 0),
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
func (srv *CoreServices) Find(identity id.Identity, name string) (*Service, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	return srv.find(identity, name)
}

func (srv *CoreServices) find(identity id.Identity, name string) (*Service, error) {
	for _, service := range srv.services {
		if service.Name() == name && service.Identity().IsEqual(identity) {
			return service, nil
		}
	}

	return nil, ErrServiceNotFound
}

func (srv *CoreServices) FindByName(name string) []*Service {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var list = make([]*Service, 0)

	for _, service := range srv.services {
		if service.Name() == name {
			list = append(list, service)
		}
	}

	return list
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
	if _, err := srv.find(identity, name); err == nil {
		return nil, ErrAlreadyRegistered
	}

	var service = newService(srv, identity, name, handler)

	// register the service
	srv.services = append(srv.services, service)

	srv.log.Infov(1, "registered %v:%s", identity, name)

	srv.events.Emit(EventServiceRegistered{
		Identity: identity,
		Name:     name,
	})

	return service, nil
}

// release removes a service from the manager
func (srv *CoreServices) release(service *Service) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	var idx = -1
	for i, s := range srv.services {
		if s == service {
			idx = i
			break
		}
	}

	if idx == -1 {
		return ErrServiceNotFound
	}

	srv.services = append(srv.services[:idx], srv.services[idx+1:]...)

	srv.log.Infov(1, "released %v:%s", service.identity, service.name)

	srv.events.Emit(EventServiceReleased{
		Identity: service.identity,
		Name:     service.name,
	})

	return nil
}
