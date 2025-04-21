package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opRemoveAssetArgs struct {
	ID     *object.ID
	Format string `query:"optional"`
}

func (mod *Module) OpRemoveAsset(ctx *astral.Context, q shell.Query, args opRemoveAssetArgs) (err error) {
	err = mod.RemoveAsset(args.ID)
	if err != nil {
		mod.log.Errorv(1, "remove asset: %v", err)
		return q.RejectWithCode(2)
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	ch.Write(&astral.Ack{})

	return
}
