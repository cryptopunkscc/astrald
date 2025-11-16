package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opPairTakeArgs struct {
	Pair astral.Nonce
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpPairTake(ctx *astral.Context, q shell.Query, args opPairTakeArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	pairEntry, ok := mod.pool.get(args.Pair)
	if !ok {
		return ch.Write(astral.NewError("pair not found"))
	}

	// NOTE: we do not allow to take a pair from another peer
	if !pairEntry.matchesPeer(q.Caller()) {
		return ch.Write(astral.NewError("peer identity does not match"))
	}

	fsm := NewPairTaker(roleTakePairResponder, ch, pairEntry)
	err = fsm.Run(ctx)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	// pair is now taken; remove it on our side from the pool.
	mod.pool.Remove(pairEntry.Nonce)

	return ch.Write(&pairEntry.EndpointPair)
}
