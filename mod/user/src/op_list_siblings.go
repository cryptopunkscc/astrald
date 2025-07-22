package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListSiblingsArgs struct {
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

func (mod *Module) OpListSiblings(ctx *astral.Context, q shell.Query, args opListSiblingsArgs) (err error) {
	ctx, cancel := ctx.WithIdentity(q.Caller()).IncludeZone(args.Zone).WithCancel()
	defer cancel()

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for _, id := range mod.listSibs() {
		err = ch.Write(id)
		if err != nil {
			return
		}
	}
	return
}
