package nat

// NOTE: might  move to mod/nat
import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opNewNatTraversal struct {
	Target string
	Out    string `query:"optional"`
}

func (mod *Module) OpNewNatTraversal(ctx *astral.Context, q shell.Query,
	args opNewNatTraversal) (err error) {
	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(4)
	}

	// acknowledge the shell query for UX completeness
	shellCh := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer shellCh.Close()

	// Start traversal by invoking the start op on the target.
	queryArgs := &opStartNatTraversal{
		Target: args.Target,
	}

	// We route the query to ourselves, which will then be forwarded to the target.
	routedQuery := query.New(ctx.Identity(), ctx.Identity(),
		nat.MethodStartNatTraversal,
		queryArgs)

	// route and get a bidirectional channel for payload exchange
	routeCh, err := query.RouteChan(ctx, mod.node, routedQuery)
	if err != nil {
		return err
	}
	defer routeCh.Close()

	// FIXME: return result of NAT traversal

	return nil
}
