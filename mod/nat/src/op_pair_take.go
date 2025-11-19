package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opPairTakeArgs struct {
	Pair astral.Nonce

	Initiate bool   `query:"optional"`
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpPairTake(ctx *astral.Context, q shell.Query, args opPairTakeArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	pair, err := mod.pool.Take(args.Pair)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	if !pair.MatchesPeer(q.Caller()) {
		return ch.Write(astral.NewError("peer identity does not match"))
	}

	if args.Initiate {
		remoteIdentity, ok := pair.RemoteIdentity(ctx.Identity())
		if !ok {
			return ch.Write(astral.NewError("remote endpoint not found"))
		}
		mod.log.Log("Pair %v: taking out of pool, starting sync with %v",
			args.Pair, remoteIdentity)

		peerCh, err := query.RouteChan(ctx.IncludeZone(astral.ZoneNetwork), mod.node,
			query.New(ctx.Identity(),
				remoteIdentity,
				nat.MethodPairTake,
				&opPairTakeArgs{
					Pair:     pair.Nonce,
					Initiate: false,
				}),
		)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		defer peerCh.Close()

		fsm := NewPairTaker(roleTakePairInitiator, peerCh, pair)
		err = fsm.Run(ctx)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return ch.Write(&pair.TraversedPortPair)
	}

	mod.log.Log("Pair %v: taking out of pool, starting sync with %v",
		args.Pair, q.Caller())
	fsm := NewPairTaker(roleTakePairResponder, ch, pair)
	err = fsm.Run(ctx)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&pair.TraversedPortPair)
}
