package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

var _ scheduler.Task = &CreateStreamTask{}
var _ nodes.CreateStreamAction = &CreateStreamTask{}

type CreateStreamTask struct {
	mod      *Module
	Target   *astral.Identity
	Endpoint exonet.Endpoint
	Info     *nodes.StreamInfo
	Err      error
}

func (m *Module) NewCreateStreamTask(target *astral.Identity, endpoint exonet.Endpoint) nodes.CreateStreamAction {
	return &CreateStreamTask{
		mod:      m,
		Target:   target,
		Endpoint: endpoint,
	}
}

func (c *CreateStreamTask) Run(ctx *astral.Context) error {
	stream, err := c.mod.peers.connectAt(ctx, c.Target, c.Endpoint)
	if err != nil {
		c.Err = err
		return err
	}

	c.Info = &nodes.StreamInfo{
		ID:             stream.id,
		LocalIdentity:  stream.LocalIdentity(),
		RemoteIdentity: stream.RemoteIdentity(),
		LocalEndpoint:  stream.LocalEndpoint(),
		RemoteEndpoint: stream.RemoteEndpoint(),
		Outbound:       astral.Bool(stream.outbound),
		Network:        astral.String8(stream.Network()),
	}

	return nil
}

func (c *CreateStreamTask) Result() (*nodes.StreamInfo, error) {
	return c.Info, c.Err
}

func (c *CreateStreamTask) String() string {
	return "nodes.create_stream"
}
