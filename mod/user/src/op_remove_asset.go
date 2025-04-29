package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opRemoveAssetArgs struct {
	ID  *object.ID
	Out string `query:"optional"`
}

func (mod *Module) OpRemoveAsset(ctx *astral.Context, q shell.Query, args opRemoveAssetArgs) (err error) {
	err = mod.RemoveAsset(args.ID)

	if err != nil {
		mod.log.Errorv(1, "remove asset: %v", err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.Write(&astral.Ack{})
}
