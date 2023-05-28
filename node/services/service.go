package services

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"time"
)

const queryQueueSize = 4

// Service represents a registered service
type Service struct {
	name         string
	queries      chan *Query
	manager      *CoreService
	identity     id.Identity
	registeredAt time.Time
}

func NewService(hub *CoreService, name string, identity id.Identity) *Service {
	return &Service{
		name:         name,
		identity:     identity,
		registeredAt: time.Now(),
		manager:      hub,
		queries:      make(chan *Query, queryQueueSize),
	}
}

// Queries returns a channel with incoming queries
func (service *Service) Queries() <-chan *Query {
	return service.queries
}

// Close closees the service
func (service *Service) Close() error {
	return service.manager.Release(service.name)
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
