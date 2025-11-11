package nodes

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

var _ scheduler.Action = &CreateStreamAction{}

type CreateStreamAction struct {
	mod      *Module
	Target   *astral.Identity
	Net      string // optional
	Endpoint string // optional in format "net:address"

	//
	Info *nodes.StreamInfo // set on success
	Err  error
}

func (m *Module) NewCreateStreamAction(target *astral.Identity, net string,
	endpoint string) nodes.CreateStreamAction {
	return &CreateStreamAction{
		mod:      m,
		Target:   target,
		Net:      net,
		Endpoint: endpoint,

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

	var endpoints chan exonet.Endpoint

	switch {
	case c.Endpoint != "":
		split := strings.SplitN(c.Endpoint, ":", 2)
		if len(split) != 2 {
			return nodes.ErrInvalidEndpointFormat
		}
		endpoint, err := c.mod.Exonet.Parse(split[0], split[1])
		if err != nil {
			return nodes.ErrEndpointParse
		}

		endpoints = make(chan exonet.Endpoint, 1)
		endpoints <- endpoint
		close(endpoints)

	case c.Net != "":
		endpoints = make(chan exonet.Endpoint, 8)
		resolve, err := c.mod.ResolveEndpoints(ctx, c.Target)
		if err != nil {
			c.mod.log.Error("resolve endpoints: %v", err)
			return nodes.ErrEndpointResolve
		}

		go func() {
			defer close(endpoints)
			for i := range resolve {
				if i.Network() == c.Net {
					endpoints <- i
				}
			}
		}()
	default:
		endpoints = make(chan exonet.Endpoint, 8)
		resolve, err := c.mod.ResolveEndpoints(ctx, c.Target)
		if err != nil {
			c.mod.log.Error("resolve endpoints: %v", err)
			return nodes.ErrEndpointResolve
		}

		go func() {
			defer close(endpoints)
			for i := range resolve {
				endpoints <- i
			}
		}()
	}

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
