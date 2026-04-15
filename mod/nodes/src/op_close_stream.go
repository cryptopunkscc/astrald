package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opCloseStreamArgs struct {
	ID  astral.Nonce
	Out string `query:"optional"`
}

// OpCloseStream closes a stream with the given id.
func (mod *Module) OpCloseStream(ctx *astral.Context, q *routing.IncomingQuery, args opCloseStreamArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	err = mod.CloseStream(args.ID)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
