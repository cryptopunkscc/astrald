package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) OpNodeOpenRelay(ctx *astral.Context, q *routing.IncomingQuery) error {
	ch := channel.New(q.AcceptRaw())
	defer ch.Close()

	return ch.Collect(func(obj astral.Object) error {
		container, ok := obj.(*nodes.QueryContainer)
		if !ok {
			return astral.NewErrUnexpectedObject(obj)
		}

		// if q.Caller() is relaying on behalf of CallerID, verify the permission
		if !container.CallerID.IsEqual(q.Caller()) {
			if !mod.Auth.Authorize(ctx, &nodes.RelayForAction{Action: auth.NewAction(q.Caller()), ForID: container.CallerID}) {
				return ch.Send(astral.NewError("unauthorized"))
			}
		}

		return mod.peers.handleRelayQuery(mod.findStreamBySessionNonce(q.Nonce()), container)
	})
}
