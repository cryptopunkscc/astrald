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
	Mod      *Module
	Target   string
	Net      string // optional
	Endpoint string // optional in format "net:address"

	//
	Info *nodes.StreamInfo // set on success
	Err  error
}

func (m *Module) NewCreateStreamAction(target string, net string,
	endpoint string) nodes.CreateStreamAction {
	return &CreateStreamAction{
		Mod:      m,
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
	target, err := c.Mod.Dir.ResolveIdentity(c.Target)
	if err != nil {
		return nodes.ErrIdentityResolve
	}

	switch {
	case c.Endpoint != "":
		split := strings.SplitN(c.Endpoint, ":", 2)
		if len(split) != 2 {
			return nodes.ErrInvalidEndpointFormat
		}
		endpoint, err := c.Mod.Exonet.Parse(split[0], split[1])
		if err != nil {
			return nodes.ErrEndpointParse
		}

		endpoints = make(chan exonet.Endpoint, 1)
		endpoints <- endpoint
		close(endpoints)

	case c.Net != "":
		endpoints = make(chan exonet.Endpoint, 8)
		resolve, err := c.Mod.ResolveEndpoints(ctx, target)
		if err != nil {
			c.Mod.log.Error("resolve endpoints: %v", err)
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
		resolve, err := c.Mod.ResolveEndpoints(ctx, target)
		if err != nil {
			c.Mod.log.Error("resolve endpoints: %v", err)
			return nodes.ErrEndpointResolve
		}

		go func() {
			defer close(endpoints)
			for i := range resolve {
				endpoints <- i
			}
		}()
	}

	s, err := c.Mod.peers.connectAtAny(ctx, target, endpoints)
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
