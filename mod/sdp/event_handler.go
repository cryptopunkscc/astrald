package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	. "github.com/cryptopunkscc/astrald/mod/sdp/api"
	"github.com/cryptopunkscc/astrald/mod/sdp/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/network"
)

type EventHandler struct {
	*Module
}

func (srv *EventHandler) Run(ctx context.Context) error {
	return events.Handle(ctx, srv.node.Events(), srv.handleLinkAdded)
}

func (srv *EventHandler) handleLinkAdded(ctx context.Context, e network.EventLinkAdded) error {
	var remoteIdentity = e.Link.RemoteIdentity()

	var conn, err = net.Route(ctx,
		srv.node.Router(),
		net.NewQuery(srv.node.Identity(), remoteIdentity, DiscoverServiceName),
	)
	if err != nil {
		srv.log.Errorv(2, "discover %s: %s", remoteIdentity, err)
		return nil
	}

	var list = make([]ServiceEntry, 0)
	for err == nil {
		err = cslq.Invoke(conn, func(msg proto.ServiceEntry) error {
			if !msg.Identity.IsEqual(remoteIdentity) {
				if srv.router != nil {
					srv.router.SetRouter(msg.Identity, remoteIdentity)
				}
			}
			list = append(list, ServiceEntry(msg))
			return nil
		})
	}

	srv.setCache(remoteIdentity, list)

	if len(list) > 0 {
		srv.log.Infov(1, "discovered %v services on %v", len(list), remoteIdentity)
		srv.events.Emit(NewEventServicesDiscovered(
			srv.log.Sprintf("%v", remoteIdentity),
			remoteIdentity,
			list,
		))
	} else {
		srv.log.Infov(1, "no services available on %v", remoteIdentity)
	}

	return nil
}
