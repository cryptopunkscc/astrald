package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSetAliasArgs struct {
	ID    *astral.Identity
	Alias string
	Out   string `query:"optional"`
}

func (mod *Module) OpSetAlias(ctx *astral.Context, q shell.Query, args opSetAliasArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	err = mod.SetAlias(args.ID, args.Alias)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
