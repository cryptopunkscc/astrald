package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opAddAssetArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (mod *Module) OpAddAsset(ctx *astral.Context, q shell.Query, args opAddAssetArgs) (err error) {
	err = mod.AddAsset(args.ID)
	if err != nil {
		mod.log.Error("error adding asset: %v", err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.Write(&astral.Ack{})
}
