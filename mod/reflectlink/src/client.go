package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/tasks"
)

type Client struct {
	*Module
}

func (mod *Client) Run(ctx context.Context) error {
	return tasks.Group(
		events.Runner(mod.node.Events(), mod.handleLinkAdded),
	).Run(ctx)
}

func (mod *Client) handleLinkAdded(ctx context.Context, event network.EventLinkAdded) error {
	// only reflect outbound links
	if !event.Link.Transport().Outbound() {
		return nil
	}

	conn, err := net.Route(ctx,
		event.Link,
		net.NewQuery(event.Link.LocalIdentity(), event.Link.RemoteIdentity(), serviceName),
	)
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

	mod.log.Logv(2, "reflected address from %v: %v %v",
		event.Link.RemoteIdentity(),
		endpoint.Network(),
		endpoint,
	)

	mod.node.Events().Emit(EventLinkReflected{Link: event.Link, Endpoint: endpoint})

	return nil
}
