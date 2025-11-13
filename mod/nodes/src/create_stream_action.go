package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

var _ scheduler.Task = &CreateStreamAction{}

type CreateStreamAction struct {
	mod       *Module
	Target    *astral.Identity
	Endpoints []exonet.Endpoint
	//
	Info *nodes.StreamInfo // set on success
	Err  error
}

func (m *Module) NewCreateStreamAction(target *astral.Identity, Endpoints []exonet.Endpoint) nodes.CreateStreamAction {
	return &CreateStreamAction{
		mod:       m,
		Target:    target,
		Endpoints: Endpoints,

		//
		Info: nil,
		Err:  nil,
	}
}

func (c *CreateStreamAction) Run(ctx *astral.Context) (err error) {
	defer func() {
		if err != nil {
			c.Err = err
		}
	}()

	endpoints := make(chan exonet.Endpoint, len(c.Endpoints))
	for _, e := range c.Endpoints {
		endpoints <- e
	}

	close(endpoints)

	s, err := c.mod.peers.connectAtAny(ctx, c.Target,
		endpoints)
	if err != nil {
		return err
	}

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

func (c *CreateStreamAction) Result() (info *nodes.StreamInfo, err error) {
	return c.Info, c.Err
}

func (c *CreateStreamAction) String() string {
	return "nodes.create_stream_action"
}
