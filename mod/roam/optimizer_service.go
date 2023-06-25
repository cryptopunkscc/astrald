package roam

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/oldlink"
	"time"
)

type OptimizerService struct {
	*Module
}

func (service *OptimizerService) Run(ctx context.Context) error {
	for event := range service.node.Events().Subscribe(ctx) {
		// skip other events
		e, ok := event.(oldlink.EventConnAdded)
		if !ok {
			continue
		}

		// skip inbound conns
		if !e.Conn.Outbound() {
			continue
		}

		// skip our own queries and silent ports
		q := e.Conn.Query()
		if q == dropServiceName || q == pickServiceName || q[0] == '.' || q[0] == ':' {
			continue
		}

		go service.optimizeConn(e.Conn)
	}

	return nil
}

func (service *OptimizerService) optimizeConn(conn *node.Conn) {
	var remoteID = conn.RemoteIdentity()
	var peer = service.node.Network().Peers().Find(remoteID)

	for {
		select {
		case <-conn.Done():
			return

		case <-time.After(time.Second):
			preferred := peer.PreferredLink()
			current := conn.Link()

			if preferred == nil {
				return
			}
			if current == preferred {
				continue
			}

			if err := service.move(conn, preferred); err != nil {
				service.log.Error("error moving connection: %s", err)
			} else {
				if current.Conns().Count() == 0 {
					current.Idle().SetTimeout(5 * time.Minute)
				}
			}
		}
	}
}

func (service *OptimizerService) move(movable *node.Conn, targetLink *oldlink.Link) error {
	var srcNet = movable.Link().Network()
	var dstNet = targetLink.Network()
	var query = movable.Query()

	service.log.Logv(1, "moving %s from %s to %s", query, srcNet, dstNet)

	moveID, err := service.pick(movable)
	if err != nil {
		return err
	}

	err = service.drop(targetLink, movable, moveID)
	if err == nil {
		service.log.Info("moved %s from %s to %s", query, srcNet, dstNet)
	}
	return err
}

func (service *OptimizerService) pick(movable *node.Conn) (int, error) {
	sourceLink := movable.Link()

	// start transfer on the source link
	server, err := net.Route(service.ctx,
		sourceLink,
		net.NewQuery(sourceLink.LocalIdentity(), sourceLink.RemoteIdentity(), pickServiceName),
	)
	if err != nil {
		return -1, err
	}
	defer server.Close()

	// write input stream id of the connection to be migrated
	if err := cslq.Encode(server, "s", movable.LocalPort()); err != nil {
		return -1, err
	}

	// read transfer id
	var moveID int
	err = cslq.Decode(server, "c", &moveID)
	if err != nil {
		return -1, err
	}

	return moveID, nil
}

func (service *OptimizerService) drop(targetLink *oldlink.Link, movable *node.Conn, moveID int) error {
	// allocate a new port on the destination link
	var newReader = oldlink.NewPortReader()
	newLocalPort, err := targetLink.Mux().BindAny(newReader)
	if err != nil {
		return err
	}

	// set connection to fall back to the new input stream
	movable.SetFallbackReader(newReader)

	// start the query
	query, err := net.Route(service.ctx,
		targetLink,
		net.NewQuery(targetLink.LocalIdentity(), targetLink.RemoteIdentity(), dropServiceName),
	)
	if err != nil {
		targetLink.Mux().Unbind(newLocalPort)
		return err
	}
	defer query.Close()

	// write move id and new input stream id
	if err := cslq.Encode(query, "cs", moveID, newLocalPort); err != nil {
		return err
	}

	// preapre the new output stream
	var newRemotePort int
	err = cslq.Decode(query, "s", &newRemotePort)
	if err != nil {
		targetLink.Mux().Unbind(newLocalPort)
		return err
	}
	newWriter := mux.NewFrameWriter(targetLink.Mux(), newRemotePort)

	// finalize the move
	if err := movable.SetWriter(newWriter); err != nil {
		return err
	}
	movable.Attach(targetLink)

	return nil
}
