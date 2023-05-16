package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type TrackedEndpoint struct {
	Identity id.Identity
	net.Endpoint
	ExpiresAt time.Time
}

type EndpointParser interface {
	Parse(network string, address string) (net.Endpoint, error)
}
