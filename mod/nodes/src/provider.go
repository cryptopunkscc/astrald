package nodes

import (
	"context"
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
)

const (
	mRelay     = "nodes.relay"
	pNonce     = "nonce"
	pSetCaller = "set_caller"
	pSetTarget = "set_target"
)

type Provider struct {
	*Module
	*routers.PathRouter

	relays sig.Map[astral.Nonce, *Relay]
}

type Relay struct {
	Caller *astral.Identity
	Target *astral.Identity
}

func NewProvider(m *Module) *Provider {
	p := &Provider{
		Module:     m,
		PathRouter: routers.NewPathRouter(m.node.Identity(), false),
	}
	p.AddRouteFunc(mRelay, p.relay)
	return p
}

func (mod *Provider) RouteQuery(ctx context.Context, q *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	return mod.PathRouter.RouteQuery(ctx, q, caller)
}

func (mod *Provider) PreprocessQuery(q *astral.Query) error {
	if r, ok := mod.relays.Delete(q.Nonce); ok {
		if !r.Target.IsZero() {
			q.Target = r.Target
		}
		if !r.Caller.IsZero() {
			q.Caller = r.Caller
		}
	}
	return nil
}

func (mod *Provider) relay(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	nonce, err := params.GetNonce(pNonce)
	if err != nil {
		mod.log.Errorv(2, "invalid relay query from %v: nonce not found", q.Caller)
		return astral.Reject()
	}

	return astral.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		r, _ := mod.relays.Set(nonce, &Relay{})

		if realCallerHex, ok := params[pSetCaller]; ok {
			realCaller, err := astral.IdentityFromString(realCallerHex)
			if err != nil {
				binary.Write(conn, binary.BigEndian, uint8(1))
				return
			}

			if !realCaller.IsEqual(q.Caller) {
				if !mod.Auth.Authorize(q.Caller, nodes.ActionRelayFor, realCaller) {
					binary.Write(conn, binary.BigEndian, uint8(1))
					return
				}
				r.Caller = realCaller
			}
		}

		if realTargetHex, ok := params[pSetTarget]; ok {
			realTarget, err := astral.IdentityFromString(realTargetHex)
			if err != nil {
				binary.Write(conn, binary.BigEndian, uint8(1))
				return
			}

			r.Target = realTarget
		}

		binary.Write(conn, binary.BigEndian, uint8(0))
	})
}
