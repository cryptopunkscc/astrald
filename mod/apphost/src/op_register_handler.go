package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opRegisterHandlerArgs struct {
	Endpoint string
	Token    astral.Nonce
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpRegisterHandler(ctx *astral.Context, q *ops.Query, args opRegisterHandlerArgs) (err error) {
	// cannot register handlers over a network
	if q.Origin() == astral.OriginNetwork {
		return q.Reject()
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// add the handler
	handler := &QueryHandler{
		Identity:  q.Caller(),
		AuthToken: args.Token,
		Endpoint:  args.Endpoint,
	}
	mod.handlers.Add(handler)

	mod.log.Logv(3, "%v registered a handler at %v", q.Caller(), args.Endpoint)

	// send ack to the client
	return ch.Send(&astral.Ack{})
}
