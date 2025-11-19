package nodes

import (
	"errors"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
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
	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(2)
	}

	var endpoints chan exonet.Endpoint

	switch {
	case args.Endpoint != "":
		split := strings.SplitN(args.Endpoint, ":", 2)
		if len(split) != 2 {
			return q.RejectWithCode(2)
		}
		endpoint, err := mod.Exonet.Parse(split[0], split[1])
		if err != nil {
			return q.RejectWithCode(3)
		}

		endpoints = make(chan exonet.Endpoint, 1)
		endpoints <- endpoint
		close(endpoints)
	case args.Net != "":
		endpoints = make(chan exonet.Endpoint, 8)
		resolve, err := mod.ResolveEndpoints(ctx, target)
		if err != nil {
			return q.RejectWithCode(4)
		}

		go func() {
			defer close(endpoints)
			for i := range resolve {
				if i.Network() == args.Net {
					endpoints <- i
				}
			}
		}()
	}
	createStreamAction := mod.NewCreateStreamAction(target, sig.ChanToArray(endpoints))
	scheduledAction, err := mod.Scheduler.Schedule(createStreamAction)
	if err != nil {
		return q.RejectWithCode(5)
	}

	// We need to accept query in case that creating stream takes longer than query timeout
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	// Wait for action or context cancllation
	select {
	case <-ctx.Done():
		return q.RejectWithCode(4)
	case <-scheduledAction.Done():
	}

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
