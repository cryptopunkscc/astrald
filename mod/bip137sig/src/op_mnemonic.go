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

	return ch.Switch(
		func(entropy *bip137sig.Entropy) error {
			words, err := bip137sig.EntropyToMnemonic(*entropy)
			if err != nil {
				return ch.Send(astral.Err(err))
			}
			phrase := strings.Join(words, " ")
			return ch.Send(astral.NewString16(phrase))
		},
		channel.PassErrors,
	)
}
