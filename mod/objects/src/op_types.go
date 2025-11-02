package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type opTypesArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpTypes(ctx *astral.Context, q shell.Query, args opTypesArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	types := mod.Blueprints().Types()
	slices.Sort(types)

	for _, name := range types {
		err = ch.Write((*astral.String)(&name))
		if err != nil {
			return
		}
	}

	return
}
