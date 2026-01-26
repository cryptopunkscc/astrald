package src

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opMnemonicToSeedArgs struct {
	Mnemonic   string
	Passphrase string `query:"optional"`

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpMnemonicToSeed(
	ctx *astral.Context,
	q *ops.Query,
	args opMnemonicToSeedArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	words := strings.Fields(args.Mnemonic)
	if len(words) == 0 {
		return ch.Send(astral.NewError("missing mnemonic"))
	}

	if _, err := bip137sig.MnemonicToEntropy(words); err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	seedBytes := bip137sig.MnemonicToSeed(words, args.Passphrase)

	return ch.Send(bip137sig.Seed{Data: seedBytes})
}
