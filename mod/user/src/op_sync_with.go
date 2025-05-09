package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSyncWithArgs struct {
	Node  *astral.Identity
	Start astral.Uint64 `query:"optional"`
	Out   string        `query:"optional"`
}

func (mod *Module) OpSyncWith(ctx *astral.Context, q shell.Query, args opSyncWithArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	err = mod.SyncAssets(ctx, args.Node)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
