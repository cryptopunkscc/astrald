package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opRemoveArgs struct {
	Object *object.ID
	Format string `query:"optional"`
}

func (mod *Module) OpRemove(ctx *astral.Context, q shell.Query, args opRemoveArgs) (err error) {
	err = mod.db.RemoveAsset(args.Object)
	if err != nil {
		mod.log.Errorv(1, "remove asset: %v", err)
		return q.RejectWithCode(2)
	}

	conn := q.Accept()
	defer conn.Close()
	
	return
}
