package relay

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const (
	ModuleName         = "relay"
	DBPrefix           = "relay__"
	ServiceName        = ".relay"
	RerouteServiceName = ".reroute"
	CertType           = "mod.relay.cert"
)

type Module interface {
	Reroute(nonce net.Nonce, router net.Router) error
	MakeCert(targetID id.Identity, relayID id.Identity, direction Direction, duration time.Duration) (object.ID, error)
	FindCerts(opts *FindOpts) ([]object.ID, error)
	Index(cert *Cert) error
	Save(cert *Cert) (object.ID, error)
	ReadCert(opts *FindOpts) ([]byte, error)
	FindExternalRelays(targetID id.Identity) ([]id.Identity, error)
	RouteVia(ctx context.Context, relayID id.Identity, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error)
	RouterFuncVia(relay id.Identity) net.RouteQueryFunc
}

type FindOpts struct {
	RelayID         id.Identity
	TargetID        id.Identity
	ExcludeRelayID  id.Identity
	ExcludeTargetID id.Identity
	Direction       Direction
	IncludeExpired  bool
}

type CertDesc struct {
	TargetID      id.Identity
	RelayID       id.Identity
	Direction     Direction
	ExpiresAt     time.Time
	ValidateError error
}

func (CertDesc) Type() string {
	return "mod.relay.cert"
}
func (d CertDesc) String() string {
	return fmt.Sprintf("Relay certificate for {{%s}}@{{%s}}", d.TargetID, d.RelayID)
}
