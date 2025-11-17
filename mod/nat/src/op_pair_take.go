package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opPairTakeArgs struct {
	Pair     astral.Nonce
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
		remoteEndpoint, ok := pair.RemoteEndpoint(ctx.Identity())
		if !ok {
			return
		}

		peerCh, err := mod.takePairQuery(ctx, remoteEndpoint.Identity, pair.Nonce)
		if err != nil {
			return
		}

		defer peerCh.Close()

		fsm := NewPairTaker(roleTakePairInitiator, peerCh, pair)
		err = fsm.Run(ctx)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		return ch.Write(&pair.TraversedEndpoints)
	}

	fsm := NewPairTaker(roleTakePairResponder, ch, pair)
	err = fsm.Run(ctx)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&pair.TraversedEndpoints)
}

func (mod *Module) takePairQuery(ctx *astral.Context, target *astral.Identity, nonce astral.Nonce) (ch *astral.Channel, err error) {
	args := &opPairTakeArgs{
		Pair: nonce,
	}

	peerQuery := query.New(ctx.Identity(), target, nat.MethodPairTake, &args)

	ch, err = query.RouteChan(
		ctx.IncludeZone(astral.ZoneNetwork),
		mod.node,
		peerQuery,
	)
	if err != nil {
		return nil, err
	}

	return ch, nil
}
