package services

const queryQueueSize = 4

// Service represents a registered service
type Service struct {
	name    string
	queries chan *Query
	manager *Manager
}

func NewService(hub *Manager, name string) *Service {
	return &Service{
		name:    name,
		manager: hub,
		queries: make(chan *Query, queryQueueSize),
	}
}

// Queries returns a channel for reading incoming queries
func (service *Service) Queries() <-chan *Query {
	return service.queries
}

// Close closees the port
func (service *Service) Close() error {
	return service.manager.release(service.name)
}

// Name returns port's name
func (service *Service) Name() string {
	return service.name
}
