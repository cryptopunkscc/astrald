package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

var _ scheduler.Task = &EnsureStreamTask{}

type EnsureStreamTask struct {
	mod        *Module
	Target     *astral.Identity
	Strategies []string
	Networks   []string
	Create     bool

	Info *nodes.StreamInfo // set on success
	Err  error
}

func (m *Module) NewEnsureStreamTask(
	target *astral.Identity,
	strategies []string,
	networks []string,
	create bool,
) nodes.EnsureStreamTask {
	return &EnsureStreamTask{
		mod:        m,
		Target:     target,
		Strategies: strategies,
		Networks:   networks,
		Create:     create,
	}
}

func (c *EnsureStreamTask) Run(ctx *astral.Context) (err error) {
	defer func() {
		if err != nil {
			c.Err = err
		}
	}()

	var opts []RetrieveLinkOption
	if c.Create {
		opts = append(opts, WithForceNew())
	}
	if len(c.Strategies) > 0 {
		opts = append(opts, WithStrategies(c.Strategies...))
	}
	if len(c.Networks) > 0 {
		opts = append(opts, WithNetworks(c.Networks...))
	}

	streamFuture := c.mod.linkPool.RetrieveLink(ctx, c.Target, opts...)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-streamFuture:
		if result.Err != nil {
			return result.Err
		}

		s := result.Stream
		c.Info = &nodes.StreamInfo{
			ID:             s.id,
			LocalIdentity:  s.LocalIdentity(),
			RemoteIdentity: s.RemoteIdentity(),
			LocalEndpoint:  s.LocalEndpoint(),
			RemoteEndpoint: s.RemoteEndpoint(),
			Outbound:       astral.Bool(s.outbound),
			Network:        astral.String8(s.Network()),
		}

		return nil
	}
}

func (c *EnsureStreamTask) Result() (info *nodes.StreamInfo, err error) {
	return c.Info, c.Err
}

func (c *EnsureStreamTask) String() string {
	return "nodes.ensure_stream"
}
