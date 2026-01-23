package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	natclient "github.com/cryptopunkscc/astrald/mod/nat/client"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opNewTraversal struct {
	Target string
	Out    string `query:"optional"`
}

func (mod *Module) OpNewTraversal(ctx *astral.Context, q shell.Query,
	args opNewTraversal) (err error) {
	_, err = mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	client := natclient.New(ctx.Identity(), astrald.Default())
	obj, err := client.StartTraversal(ctx, args.Target)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(obj)
}
