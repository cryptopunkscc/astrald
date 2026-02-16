package nodes

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	if e.Endpoint == nil {
		return 0, errors.New("nil endpoint")
	}

	n, err = astral.Encode(w, e.Endpoint)
	if err != nil {
		return
	}

	hasTTL := e.TTL != nil
	if err = binary.Write(w, astral.ByteOrder, hasTTL); err != nil {
		return
	}
	n++

	if hasTTL {
		if err = binary.Write(w, astral.ByteOrder, *e.TTL); err != nil {
			return
		}
		n += 4
	}

	return
}

func (e *ResolvedEndpoint) ReadFrom(r io.Reader) (n int64, err error) {
	obj, m, err := astral.Decode(r)
	n += m
	if err != nil {
		return
	}

	ep, ok := obj.(exonet.Endpoint)
	if !ok {
		return n, fmt.Errorf("object is not an exonet.Endpoint")
	}
	e.Endpoint = ep

	var hasTTL bool
	if err = binary.Read(r, astral.ByteOrder, &hasTTL); err != nil {
		return
	}
	n++

	if hasTTL {
		var ttl uint32
		if err = binary.Read(r, astral.ByteOrder, &ttl); err != nil {
			return
		}
		n += 4
		e.TTL = &ttl
	} else {
		e.TTL = nil
	}

	return
}

// exonet.Endpoint

func (e *ResolvedEndpoint) Network() string { return e.Endpoint.Network() }
func (e *ResolvedEndpoint) Address() string { return e.Endpoint.Address() }
func (e *ResolvedEndpoint) Pack() []byte    { return e.Endpoint.Pack() }

func init() {
	_ = astral.Add(&ResolvedEndpoint{})
}
