package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opAliasMapArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpAliasMap(ctx *astral.Context, q shell.Query, args opAliasMapArgs) (err error) {
	ch := channel.New(q.Accept(), channel.InFmt(args.In), channel.OutFmt(args.Out))
	defer ch.Close()

	return ch.Write(mod.AliasMap())
}
