package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opAssetsArgs struct {
	Format string `query:"optional"`
}

func (mod *Module) OpAssets(ctx *astral.Context, q shell.Query, args opAssetsArgs) (err error) {
	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	for _, asset := range mod.Assets() {
		err = ch.Write(asset)
		if err != nil {
			return err
		}
	}

	return
}
