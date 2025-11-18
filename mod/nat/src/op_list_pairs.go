package nat

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListPairsArgs struct {
	With astral.String `query:"optional"`
	Out  string        `query:"optional"`
}

func (mod *Module) OpListPairs(ctx *astral.Context, q shell.Query,
	args opListPairsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	pairs := mod.pool.GetAll()
	fmt.Println("pairs length listing: ", len(pairs))
	for _, pair := range pairs {
		mod.log.Info("pair: %v", pair)
		if args.With != "" {
			target, err := mod.Dir.ResolveIdentity(string(args.With))
			if err != nil {
				return ch.Write(astral.NewError(err.Error()))
			}

			if !pair.MatchesPeer(target) {
				continue
			}
		}

		err = ch.Write(&pair.TraversedPortPair)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
	}

	return ch.Write(&astral.Ack{})
}
