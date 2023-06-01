package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/network"
)

type EventHandler struct {
	*Module
}

func (srv *EventHandler) Run(ctx context.Context) error {
	return event.Handle(ctx, srv.node.Events(), func(e network.EventPeerLinked) error {
		srv.handlePeerLinked(ctx, e)

		return nil
	})
}

func (srv *EventHandler) handlePeerLinked(ctx context.Context, e network.EventPeerLinked) error {
	conn, err := srv.node.Network().Query(ctx, e.Peer.Identity(), discoverServiceName)
	if err != nil {
		srv.log.Errorv(2, "discover %s: %s", e.Peer.Identity(), err)
		return err
	}

	var list = make([]proto.ServiceEntry, 0)
	for err == nil {
		err = cslq.Invoke(conn, func(msg proto.ServiceEntry) error {
			list = append(list, msg)
			return nil
		})
	}

	if len(list) > 0 {
		srv.log.Infov(2, "discover %s: %v services available", e.Peer.Identity(), len(list))
		srv.events.Emit(EventPeerServices{
			Identity: e.Peer.Identity(),
			Services: list,
		})
	}

	return nil
}
