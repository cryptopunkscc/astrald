package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opGetAliasArgs struct {
	ID  *astral.Identity
	Out string `query:"optional"`
}

func (mod *Module) OpGetAlias(ctx *astral.Context, q shell.Query, args opGetAliasArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	alias, err := mod.GetAlias(args.ID)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	
	return ch.Write((*astral.String8)(&alias))
}
