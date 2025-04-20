package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opAddArgs struct {
	Object *object.ID
	Format string `query:"optional"`
}

func (mod *Module) OpAdd(ctx *astral.Context, q shell.Query, args opAddArgs) (err error) {
	nonce, err := mod.db.AddAsset(args.Object, false)
	if err != nil {
		mod.log.Error("error adding asset: %v", err)
		return q.RejectWithCode(2)
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	return ch.Write(&nonce)
}
