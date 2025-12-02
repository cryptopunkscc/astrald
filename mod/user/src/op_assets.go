package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opAssetsArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpAssets(ctx *astral.Context, q shell.Query, args opAssetsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for _, asset := range mod.Assets() {
		err = ch.Write(asset)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
	}

	return ch.Write(&astral.EOS{})
}
