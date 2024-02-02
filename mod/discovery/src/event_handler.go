package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/network"
)

type EventHandler struct {
	*Module
}

func (srv *EventHandler) Run(ctx context.Context) error {
	return events.Handle(ctx, srv.node.Events(), func(e network.EventLinkAdded) error {
		return srv.handleLinkAdded(ctx, e)
	})
}

func (srv *EventHandler) handleLinkAdded(ctx context.Context, e network.EventLinkAdded) error {
	var remoteIdentity = e.Link.RemoteIdentity()

	info, err := srv.DiscoverRemote(ctx, remoteIdentity, srv.node.Identity())
	if err != nil {
		srv.log.Errorv(2, "discover %s: %s", remoteIdentity, err)
		return nil
	}

	srv.log.Infov(1,
		"discovered %v data items and %v services on %v",
		len(info.Data),
		len(info.Services),
		remoteIdentity,
	)

	srv.events.Emit(discovery.NewEventDiscovered(
		srv.log.Sprintf("%v", remoteIdentity),
		remoteIdentity,
		info,
	))

	return nil
}
