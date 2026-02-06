package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

var _ scheduler.Task = &EnsureStreamTask{}

type EnsureStreamTask struct { // note: maybe ensure link task?
	mod      *Module
	Target   *astral.Identity
	Endpoint exonet.Endpoint
	Network  *string
	Create   bool

	Info *nodes.StreamInfo // set on success
	Err  error
}

func (m *Module) NewEnsureStreamTask(target *astral.Identity, endpoint exonet.Endpoint, network *string, create bool) nodes.EnsureStreamTask {
	return &EnsureStreamTask{
		mod:      m,
		Target:   target,
		Endpoint: endpoint,
		Network:  network,
		Create:   create,
	}
}

func (c *EnsureStreamTask) Run(ctx *astral.Context) (err error) {
	defer func() {
		if err != nil {
			c.Err = err
		}
	}()

	var retrieveLinkOptions []RetrieveLinkOption
	if c.Create {
		retrieveLinkOptions = append(retrieveLinkOptions, WithForceNew())
	}

	switch {
	case c.Endpoint != nil:
		retrieveLinkOptions = append(retrieveLinkOptions, WithEndpoints(c.Endpoint))
	case c.Network != nil:
		retrieveLinkOptions = append(retrieveLinkOptions, WithIncludeNetworks(*c.Network))
	}

	streamFuture := c.mod.linkPool.RetrieveLink(ctx, c.Target, retrieveLinkOptions...)
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
