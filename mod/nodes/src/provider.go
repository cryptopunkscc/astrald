package nodes

import (
	"context"
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
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

type relayArgs struct {
	Nonce     astral.Nonce
	SetCaller *astral.Identity `query:"optional"`
	SetTarget *astral.Identity `query:"optional"`
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
	var args relayArgs

	_, err := query.ParseTo(q.Query, &args)
	if err != nil {
		mod.log.Errorv(2, "%v -> relay: invalid arguments: %v", q.Caller, err)
		return astral.Reject()
	}

	return astral.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		r, _ := mod.relays.Set(args.Nonce, &Relay{})

		if !args.SetCaller.IsZero() {
			if !args.SetCaller.IsEqual(q.Caller) {
				if !mod.Auth.Authorize(q.Caller, nodes.ActionRelayFor, args.SetCaller) {
					binary.Write(conn, binary.BigEndian, uint8(1))
					return
				}
				r.Caller = args.SetCaller
			}
		}

		if !args.SetTarget.IsZero() {
			r.Target = args.SetTarget
		}

		binary.Write(conn, binary.BigEndian, uint8(0))
	})
}
