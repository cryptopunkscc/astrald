package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opResolveArgs struct {
	Name string
	Out  string `query:"optional"`
}

func (mod *Module) OpResolve(ctx *astral.Context, q shell.Query, args opResolveArgs) (err error) {
	id, err := mod.ResolveIdentity(args.Name)
	if err != nil {
		return q.RejectWithCode(8)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.Write(id)
}
