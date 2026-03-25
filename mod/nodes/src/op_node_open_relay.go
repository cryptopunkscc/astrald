package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) OpNodeOpenRelay(_ *astral.Context, q *ops.Query) error {
	ch := channel.New(q.Accept())
	defer ch.Close()

	return ch.Collect(func(obj astral.Object) error {
		container, ok := obj.(*nodes.QueryContainer)
		if !ok {
			return astral.NewErrUnexpectedObject(obj)
		}

		if !container.CallerID.IsEqual(q.Caller()) {
			if !mod.Auth.Authorize(q.Caller(), nodes.ActionRelayFor, container.CallerID) {
				return ch.Send(astral.NewError("unauthorized"))
			}
		}

		err := mod.peers.handleRelayQuery(container, q.Nonce)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(&astral.Ack{})
	})
}
