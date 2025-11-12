package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ scheduler.Action = &CreateStreamAction{}

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

	s, err := c.mod.peers.connectAtAny(ctx, c.Target, sig.ArrayToChan(c.Endpoints))
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
