package nodes

import (
	"errors"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type opNewStreamArgs struct {
	Target   string
	Net      string `query:"optional"`
	Endpoint string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpNewStream(ctx *astral.Context, q *ops.Query, args opNewStreamArgs) (err error) {
	target, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return q.RejectWithCode(2)
	}

	var endpoint exonet.Endpoint
	if args.Endpoint != "" {
		split := strings.SplitN(args.Endpoint, ":", 2)
		if len(split) == 2 {
			var parseErr error
			endpoint, parseErr = mod.Exonet.Parse(split[0], split[1])
			if parseErr != nil {
				return q.RejectWithCode(3)
			}
		}
	}

	var network *string
	if args.Net != "" {
		network = &args.Net
	}

	task := mod.NewCreateStreamTask(target, endpoint, network)
	scheduledTask, err := mod.Scheduler.Schedule(task)
	if err != nil {
		return q.RejectWithCode(5)
	}

	// We need to accept query in case that creating stream takes longer than query timeout
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	// Wait for task or context cancellation
	select {
	case <-ctx.Done():
		return q.RejectWithCode(4)
	case <-scheduledTask.Done():
	}

	info, err := task.Result()
	switch {
	case err == nil:
		return ch.Send(info)
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
