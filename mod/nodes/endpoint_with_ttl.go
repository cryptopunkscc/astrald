package nodes

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

var _ exonet.Endpoint = &EndpointWithTTL{}
var _ astral.Object = &EndpointWithTTL{}

// EndpointWithTTL pairs an exonet.Endpoint with an optional TTL in seconds.
// A nil TTL means the endpoint does not expire.
type EndpointWithTTL struct {
	Endpoint exonet.Endpoint
	TTL      *uint32 // seconds, nil = no expiry
}

func NewEndpointWithTTL(endpoint exonet.Endpoint, ttl ...time.Duration) *EndpointWithTTL {
	re := &EndpointWithTTL{Endpoint: endpoint}
	if len(ttl) > 0 {
		var secs uint32 = 0
		for _, d := range ttl {
			secs += uint32(d.Seconds())
		}

		re.TTL = &secs
	}

	return re
}

func (EndpointWithTTL) ObjectType() string {
	return "mod.nodes.endpoint_with_ttl"
}

func (e EndpointWithTTL) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *EndpointWithTTL) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

func (e EndpointWithTTL) MarshalJSON() ([]byte, error) {
	endpointJSON, err := json.Marshal(e.Endpoint)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Endpoint astral.JSONAdapter
		TTL      *uint32 `json:",omitempty"`
	}{
		Endpoint: astral.JSONAdapter{Type: e.Endpoint.ObjectType(), Object: endpointJSON},
		TTL:      e.TTL,
	})
}

func (e *EndpointWithTTL) UnmarshalJSON(data []byte) error {
	var v struct {
		Endpoint astral.JSONAdapter
		TTL      *uint32 `json:",omitempty"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	obj := astral.New(v.Endpoint.Type)
	if obj == nil {
		return astral.NewErrBlueprintNotFound(v.Endpoint.Type)
	}
	if err := json.Unmarshal(v.Endpoint.Object, obj); err != nil {
		return err
	}
	ep, ok := obj.(exonet.Endpoint)
	if !ok {
		return errors.New("EndpointWithTTL: not an exonet.Endpoint")
	}
	e.Endpoint = ep
	e.TTL = v.TTL
	return nil
}

// exonet.Endpoint

func (e *EndpointWithTTL) Network() string { return e.Endpoint.Network() }
func (e *EndpointWithTTL) Address() string { return e.Endpoint.Address() }
func (e *EndpointWithTTL) Pack() []byte    { return e.Endpoint.Pack() }

func init() {
	_ = astral.Add(&EndpointWithTTL{})
}
