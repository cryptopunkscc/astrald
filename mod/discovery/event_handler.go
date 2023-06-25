package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
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
		srv.node.Network(),
		net.NewQuery(srv.node.Identity(), remoteIdentity, discoverServiceName),
	)
	if err != nil {
		srv.log.Errorv(2, "discover %s: %s", remoteIdentity, err)
		return nil
	}

	var list = make([]ServiceEntry, 0)
	for err == nil {
		err = cslq.Invoke(conn, func(msg rpc.ServiceEntry) error {
			list = append(list, ServiceEntry(msg))
			return nil
		})
	}

	srv.setCache(remoteIdentity, list)

	if len(list) > 0 {
		srv.log.Infov(2, "discover %s: %v services available", remoteIdentity, len(list))
		srv.events.Emit(EventServicesDiscovered{
			identityName: srv.log.Sprintf("%v", remoteIdentity),
			Identity:     remoteIdentity,
			Services:     list,
		})
	} else {
		srv.log.Infov(2, "discover %s: no services available", remoteIdentity)
	}

	return nil
}
