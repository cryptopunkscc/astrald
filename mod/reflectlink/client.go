package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Client struct {
	*Module
}

func (mod *Client) Run(ctx context.Context) error {
	return event.Handle(ctx, mod.node.Events(), mod.handleLinkEstablished)
}

func (mod *Client) handleLinkEstablished(ctx context.Context, event link.EventLinkEstablished) error {
	conn, err := event.Link.Query(ctx, serviceName)
	if err != nil {
		return err
	}
	defer conn.Close()

	var ref proto.Reflection

	err = json.NewDecoder(conn).Decode(&ref)
	if err != nil {
		return err
	}

	endpoint, err := mod.node.Infra().Parse(ref.RemoteEndpoint.Network, ref.RemoteEndpoint.Address)
	if err != nil {
		return err
	}

	mod.node.Events().Emit(EventLinkReflected{Link: event.Link, Endpoint: endpoint})

	return nil
}
