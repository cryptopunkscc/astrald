package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/network"
)

type EventHandler struct {
	*Module
}

func (srv *EventHandler) Run(ctx context.Context) error {
	return events.Handle(ctx, srv.node.Events(), srv.handlePeerLinked)
}

func (srv *EventHandler) handlePeerLinked(ctx context.Context, e network.EventPeerLinked) error {
	var identity = e.Peer.Identity()
	var conn, err = srv.node.Network().Query(ctx, identity, discoverServiceName)
	if err != nil {
		srv.log.Errorv(2, "discover %s: %s", identity, err)
		return err
	}

	var list = make([]ServiceEntry, 0)
	for err == nil {
		err = cslq.Invoke(conn, func(msg rpc.ServiceEntry) error {
			list = append(list, ServiceEntry(msg))
			return nil
		})
	}

	srv.setCache(identity, list)

	if len(list) > 0 {
		srv.log.Infov(2, "discover %s: %v services available", identity, len(list))
		srv.events.Emit(EventServicesDiscovered{
			identityName: srv.log.Sprintf("%v", identity),
			Identity:     identity,
			Services:     list,
		})
	}

	return nil
}
