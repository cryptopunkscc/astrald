package nodes

import (
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type ObservedEndpointMessage struct {
	Endpoint exonet.Endpoint
}

func (ObservedEndpointMessage) ObjectType() string {
	return "mod.nodes.observed_endpoint_message"
}

func (e ObservedEndpointMessage) WriteTo(w io.Writer) (n int64, err error) {
	if e.Endpoint == nil {
		return 0, errors.New("nil endpoint")
	}
	return astral.Write(w, e.Endpoint)
}

func (e *ObservedEndpointMessage) ReadFrom(r io.Reader) (n int64, err error) {
	bp := astral.ExtractBlueprints(r)

	obj, m, err := bp.Read(r, false)
	n += m
	if err != nil {
		return
	}

	if ep, ok := obj.(exonet.Endpoint); ok {
		e.Endpoint = ep
		return
	}

	if raw, ok := obj.(*astral.RawObject); ok {
		// Try to refine raw into a concrete type using blueprints
		if refined, refineErr := bp.Refine(raw); refineErr == nil {
			if ep, ok := refined.(exonet.Endpoint); ok {
				e.Endpoint = ep
				return n, nil
			}
		}
	}

	return n, fmt.Errorf("object is not an exonet.Endpoint")
}

func init() {
	_ = astral.DefaultBlueprints.Add(&ObservedEndpointMessage{})
}
