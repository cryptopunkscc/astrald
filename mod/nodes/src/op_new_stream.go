package nodes

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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
	case <-w.Wait():
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

	info, err := createStreamAction.Result()
	switch {
	case err == nil:
		return ch.Write(info)
	case errors.Is(err, nodes.ErrInvalidEndpointFormat):
		return q.RejectWithCode(2)
	case errors.Is(err, nodes.ErrEndpointParse):
		return q.RejectWithCode(3)
	case errors.Is(err, nodes.ErrIdentityResolve), errors.Is(err, nodes.ErrEndpointResolve):
		return q.RejectWithCode(4)
	default:
		return err
	}
}
