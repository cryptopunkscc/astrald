package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	nodes2 "github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
)

type opNewStreamArgs struct {
	Target   string
	Net      string `query:"optional"`
	Endpoint string `query:"optional"`
	Out      string `query:"optional"`
}

// OpNewStream now delegates to a scheduled action and waits for completion.
func (mod *Module) OpNewStream(ctx *astral.Context, q shell.Query, args opNewStreamArgs) (err error) {
	createStreamAction := mod.NewCreateStreamAction(args.Target, args.Net, args.Endpoint)
	w := mod.Scheduler.Schedule(ctx, createStreamAction)

	// Wait for action or context cancellation
	select {
	case <-ctx.Done():
		return q.RejectWithCode(4)
	case <-w.Done():
	}

	if err := w.Error(); err != nil {
		switch err {
		case nodes2.ErrInvalidEndpointFormat:
			return q.RejectWithCode(2)
		case nodes2.ErrEndpointParse:
			return q.RejectWithCode(3)
		case nodes2.ErrIdentityResolve, nodes2.ErrEndpointResolve:
			return q.RejectWithCode(4)
		default:
			ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
			defer ch.Close()
			return ch.Write(astral.NewError(err.Error()))
		}
	}

	createStreamAction := mod.NewCreateStreamAction(target, sig.ChanToArray(endpoints))
	scheduledAction, err := mod.Scheduler.Schedule(ctx, createStreamAction)
	if err != nil {
		return q.RejectWithCode(5)
	}

	// Wait for action or context cancellation
	select {
	case <-ctx.Done():
		return q.RejectWithCode(4)
	case <-scheduledAction.Done():
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.Write(createStreamAction.Info)
}
