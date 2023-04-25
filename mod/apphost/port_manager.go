package apphost

import (
	"github.com/cryptopunkscc/astrald/node/services"
	"sync"
)

type PortManager struct {
	ports map[string]PortEntry
	mu    sync.Mutex
}

func NewPortManager() *PortManager {
	return &PortManager{
		ports: make(map[string]PortEntry),
	}
}

func (m *PortManager) GetPort(portName string) *services.Service {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkPort(portName)

	if port, found := m.ports[portName]; found {
		return port.port
	}
	return nil
}

func (m *PortManager) AddPort(port *services.Service, target string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	portName := port.Name()

	if _, found := m.ports[portName]; found {
		return services.ErrAlreadyRegistered
	}

	m.ports[portName] = PortEntry{
		port:   port,
		target: target,
	}

	return nil
}

// checkPort removes a port from the manager if checkTarget fails
func (m *PortManager) checkPort(portName string) {
	if entry, found := m.ports[portName]; found {
		if entry.checkTarget() {
			return
		}
		entry.port.Close()
		delete(m.ports, portName)
	}
}
