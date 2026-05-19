package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type opRegisterDescriberArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpRegisterDescriber(ctx *astral.Context, q *routing.IncomingQuery, args opRegisterDescriberArgs) error {
	// Keep this local for now; extract shared app registration validation once the API settles.
	if q.Origin() == astral.OriginNetwork {
		return q.Reject()
	}

	id := q.Caller()
	var err error
	switch {
	case id == nil || id.IsZero():
		err = objects.ErrInvalidSourceIdentity
	case id.IsEqual(mod.node.Identity()):
		err = objects.ErrAppRegistrationSelf
	}

	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = mod.AddDescriber(NewAppDescriber(mod, id))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
