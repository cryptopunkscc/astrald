package src

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opMnemonicArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpMnemonic(
	ctx *astral.Context,
	q *ops.Query,
	args opMnemonicArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *bip137sig.Entropy:
			words, err := bip137sig.EntropyToMnemonic(*object)
			if err != nil {
				ch.Send(astral.Err(err))
				return
			}
			phrase := strings.Join(words, " ")
			ch.Send(astral.NewString16(phrase))

		default:
			ch.Send(astral.NewErrUnexpectedObject(object))
		}
	})
}
