package services

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"sync/atomic"
	"time"
)

var _ net.Router = &Service{}

// Service represents a registered service
type Service struct {
	net.Router
	name         string
	manager      *CoreServices
	identity     id.Identity
	registeredAt time.Time
	done         chan struct{}
	closed       atomic.Bool
}

func newService(hub *CoreServices, identity id.Identity, name string, router net.Router) *Service {
	return &Service{
		manager:      hub,
		identity:     identity,
		name:         name,
		Router:       router,
		registeredAt: time.Now(),
		done:         make(chan struct{}),
	}
}

// Close closees the service
func (service *Service) Close() error {
	if service.closed.CompareAndSwap(false, true) {
		service.manager.release(service)
		close(service.done)
		return nil
	}
	return errors.New("already done")
}

func (service *Service) Done() <-chan struct{} {
	return service.done
}

// Name returns service's name
func (service *Service) Name() string {
	return service.name
}

// Identity returns service's identity
func (service *Service) Identity() id.Identity {
	return service.identity
}

// RegisteredAt returns service's registration time
func (service *Service) RegisteredAt() time.Time {
	return service.registeredAt
}
