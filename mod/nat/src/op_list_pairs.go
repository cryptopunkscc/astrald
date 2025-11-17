package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListPairsArgs struct {
	With *astral.Identity `query:"optional"`
	Out  string           `query:"optional"`
}

func (mod *Module) OpListPairs(ctx *astral.Context, q shell.Query,
	args opListPairsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	pairs := mod.pool.GetAll()

	for _, pair := range pairs {
		if args.With != nil {
			if !pair.MatchesPeer(args.With) {
				continue
			}
		}

		err = ch.Write(&pair.TraversedEndpoints)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
	}

	return nil
}
