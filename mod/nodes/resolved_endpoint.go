package nodes

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Endpoint = &ResolvedEndpoint{}
var _ astral.Object = &ResolvedEndpoint{}

// ResolvedEndpoint pairs an exonet.Endpoint with an optional TTL in seconds.
// A nil TTL means the endpoint does not expire.
type ResolvedEndpoint struct {
	Endpoint exonet.Endpoint
	TTL      *uint32 // seconds, nil = no expiry
}

func NewResolvedEndpoint(endpoint exonet.Endpoint, ttl ...time.Duration) *ResolvedEndpoint {
	re := &ResolvedEndpoint{Endpoint: endpoint}
	if len(ttl) > 0 {
		var secs uint32 = 0
		for _, d := range ttl {
			secs += uint32(d.Seconds())
		}

		re.TTL = &secs
	}

	return re
}

func (ResolvedEndpoint) ObjectType() string {
	return "mod.nodes.resolved_endpoint"
}

func (e ResolvedEndpoint) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *ResolvedEndpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

// exonet.Endpoint

func (e *ResolvedEndpoint) Network() string { return e.Endpoint.Network() }
func (e *ResolvedEndpoint) Address() string { return e.Endpoint.Address() }
func (e *ResolvedEndpoint) Pack() []byte    { return e.Endpoint.Pack() }

func init() {
	_ = astral.Add(&ResolvedEndpoint{})
}
