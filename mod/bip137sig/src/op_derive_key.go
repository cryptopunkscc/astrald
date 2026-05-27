package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opDeriveKeyArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpDeriveKey(
	ctx *astral.Context,
	q *routing.IncomingQuery,
	args opDeriveKeyArgs,
) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	object, err := ch.Receive()
	if err != nil {
		return err
	}

	seed, ok := object.(*bip137sig.Seed)
	if !ok {
		return ch.Send(astral.NewErrUnexpectedObject(object))
	}

	privateKey, err := mod.DeriveKey(*seed, args.Path)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&privateKey)
}
