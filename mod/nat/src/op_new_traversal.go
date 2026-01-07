package nat

// NOTE: might  move to mod/nat
import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
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

	queryArgs := &opNewTraversal{
		Target: args.Target,
	}

	routedQuery := query.New(ctx.Identity(), ctx.Identity(),
		nat.MethodStartNatTraversal,
		queryArgs)

	routeCh, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}
	defer routeCh.Close()

	obj, err := routeCh.Receive()
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(obj)
}
