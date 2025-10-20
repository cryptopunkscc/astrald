package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListPairsArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpListPairs(ctx *astral.Context, q shell.Query,
	args opListPairsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	pairs := mod.traversedPairs
	for _, pair := range pairs {
		err = ch.Write(&pair)
		if err != nil {
			return err
		}
	}

	return nil
}
