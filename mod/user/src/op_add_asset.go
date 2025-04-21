package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opAddAssetArgs struct {
	ID     *object.ID
	Format string `query:"optional"`
}

func (mod *Module) OpAddAsset(ctx *astral.Context, q shell.Query, args opAddAssetArgs) (err error) {
	err = mod.AddAsset(args.ID)
	if err != nil {
		mod.log.Error("error adding asset: %v", err)
		return q.RejectWithCode(2)
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	return ch.Write(&astral.Ack{})
}
