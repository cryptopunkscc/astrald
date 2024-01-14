package relay

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const (
	ModuleName         = "relay"
	RelayServiceName   = ".relay"
	RerouteServiceName = ".reroute"
	RelayCertType      = "cert.router.relay"
)

type Module interface {
	Reroute(nonce net.Nonce, router net.Router) error
	MakeCert(targetID id.Identity, relayID id.Identity, direction Direction, duration time.Duration) (data.ID, error)
	FindCerts(opts *FindOpts) ([]data.ID, error)
	ReadCert(opts *FindOpts) ([]byte, error)
}

type FindOpts struct {
	RelayID         id.Identity
	TargetID        id.Identity
	ExcludeRelayID  id.Identity
	ExcludeTargetID id.Identity
	Direction       Direction
	IncludeExpired  bool
}
