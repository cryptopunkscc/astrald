package nat

// NOTE: might  move to mod/nat
import (
	"github.com/cryptopunkscc/astrald/astral"
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

	// acknowledge the shell query for UX completeness
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	// StartKeepAlive traversal by invoking the start op on the target.
	queryArgs := &opNewTraversal{
		Target: args.Target,
	}

	// We route the query to ourselves, which will then be forwarded to the target.
	routedQuery := query.New(ctx.Identity(), ctx.Identity(),
		nat.MethodStartNatTraversal,
		queryArgs)

	// route and Get a bidirectional channel for payload exchange
	routeCh, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	defer routeCh.Close()

	obj, err := routeCh.Read()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(obj)
}
