package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (rw io.WriteCloser, err error) {
	if s, ok := q.Extra.Get("origin"); ok && s == "network" {
		return query.RouteNotFound(mod)
	}

	switch {
	case mod.IsPeer(q.Target):
	case len(mod.Endpoints(q.Target)) > 0:
	case q.Target.IsEqual(mod.node.Identity()):
		return query.RouteNotFound(mod)
	default:
		return query.RouteNotFound(mod)
	}

	var relayID = q.Target
	var callerProof astral.Object

	useRelay := false

	if !q.Caller.IsEqual(mod.node.Identity()) {
		useRelay = true
		if v, ok := q.Extra.Get(nodes.ExtraCallerProof); ok {
			callerProof = v.(astral.Object)
		}
	}

	if v, ok := q.Extra.Get(nodes.ExtraRelayVia); ok {
		relayID = v.(*astral.Identity)
		useRelay = true
	}

	if useRelay {
		if callerProof != nil {
			actx := astral.NewContext(ctx).WithIdentity(mod.node.Identity())
			err = mod.Objects.Push(actx, relayID, callerProof)
			if err != nil {
				mod.log.Errorv(1, "cannot push proof: %v", err)
			}
		}

		err = mod.on(relayID).Relay(ctx, q.Nonce, q.Caller, q.Target)
		if err != nil {
			return query.RouteNotFound(mod, err)
		}

		if !q.Target.IsEqual(relayID) {
			q = &astral.Query{
				Nonce:  q.Nonce,
				Caller: q.Caller,
				Target: relayID,
				Query:  q.Query,
				Extra:  *q.Extra.Copy(),
			}
		}
	}

	return mod.peers.RouteQuery(ctx, q, w)
}
