package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListPairsArgs struct {
	With astral.String `query:"optional"`
	Out  string        `query:"optional"`
}

func (mod *Module) OpListPairs(ctx *astral.Context, q shell.Query,
	args opListPairsArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	pairs := mod.pool.GetAll()
	for _, pair := range pairs {
		if args.With != "" {
			target, err := mod.Dir.ResolveIdentity(string(args.With))
			if err != nil {
				return ch.Send(astral.NewError(err.Error()))
			}

			if !pair.MatchesPeer(target) {
				continue
			}
		}

		err = ch.Send(&pair.TraversedPortPair)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
