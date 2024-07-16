package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
)

const ModuleName = "discovery"
const DiscoverServiceName = ".discover"

type Module interface {
	AddServiceDiscoverer(ServiceDiscoverer) error
	RemoveServiceDiscoverer(ServiceDiscoverer) error
	AddDataDiscoverer(DataDiscoverer) error
	RemoveDataDiscoverer(DataDiscoverer) error
}

type ServiceDiscoverer interface {
	DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]Service, error)
}

type DataDiscoverer interface {
	DiscoverData(ctx context.Context, caller id.Identity, origin string) ([][]byte, error)
}

// Info contains a full discovery response
type Info struct {
	Data     []Data
	Services []Service
}

// Data respresents a single data entry in discovery response
type Data struct {
	Identity id.Identity
	Bytes    []byte
}

// Service represents a single service entry in discovery response
type Service struct {
	Identity id.Identity
	Name     string
	Type     string
	Extra    []byte
}
