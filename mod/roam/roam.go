package roam

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/services"
	"sync"
	"time"
)

const (
	portPick = "roam.pick"
	portDrop = "roam.drop"
)

type Module struct {
	node  node.Node
	moves map[int]*link.Conn
	log   *log.Logger
}

func (mod *Module) Run(ctx context.Context) error {
	mod.moves = make(map[int]*link.Conn)

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		mod.servePick(ctx)
		wg.Done()
	}()

	go func() {
		mod.serveDrop(ctx)
		wg.Done()
	}()

	go func() {
		mod.monitorConnections(ctx)
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (mod *Module) monitorConnections(ctx context.Context) {
	for event := range mod.node.Events().Subscribe(ctx) {
		// skip other events
		e, ok := event.(link.EventConnEstablished)
		if !ok {
			continue
		}

		// skip inbound conns
		if !e.Conn.Outbound() {
			continue
		}

		// skip our own queries and silent ports
		q := e.Conn.Query()
		if q == portDrop || q == portPick || q[0] == '.' || q[0] == ':' {
			continue
		}

		go mod.optimizeConn(e.Conn)
	}
}

func (mod *Module) servePick(ctx context.Context) error {
	var queries = services.NewQueryChan(4)
	service, err := mod.node.Services().Register(ctx, mod.node.Identity(), portPick, queries.Push)
	if err != nil {
		return err
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for query := range queries {
		// skip local queries
		if query.Origin() == services.OriginLocal {
			query.Reject()
			continue
		}

		client, err := query.Accept()
		if err != nil {
			continue
		}

		var remotePort int

		// read remote port of the connection to pick
		cslq.Decode(client, "s", &remotePort)

		// find the connection
		c := client.Link().Conns().FindByRemotePort(remotePort)
		if c == nil {
			client.Close()
			continue
		}

		// allocate a new move for the connection
		moveID := mod.allocMove(c)
		if moveID != -1 {
			cslq.Encode(client, "c", moveID)
		}

		client.Close()
	}

	return nil
}

func (mod *Module) serveDrop(ctx context.Context) {
	var queries = services.NewQueryChan(4)

	service, err := mod.node.Services().Register(ctx, mod.node.Identity(), portDrop, queries.Push)
	if err != nil {
		return
	}

	go func() {
		<-service.Done()
		close(queries)
	}()

	for query := range queries {
		// skip local queries
		if query.Origin() == services.OriginLocal {
			query.Reject()
			continue
		}

		client, _ := query.Accept()

		var targetLink = client.Link()
		var moveID, newRemotePort int

		if err := cslq.Decode(client, "cs", &moveID, &newRemotePort); err != nil {
			client.Close()
			continue
		}

		movable, found := mod.moves[moveID]
		if !found {
			client.Close()
			continue
		}

		// allocate a new input stream and write its id
		var newReader = link.NewPortReader()
		newLocalPort, err := targetLink.Mux().BindAny(newReader)
		if err != nil {
			client.Close()
			continue
		}

		movable.SetFallbackReader(newReader)

		cslq.Encode(client, "s", newLocalPort)

		// replace the output stream and finalize the move
		newWriter := mux.NewFrameWriter(targetLink.Mux(), newRemotePort)
		movable.SetWriter(newWriter)
		movable.Attach(targetLink)

		client.Close()
	}
}

func (mod *Module) optimizeConn(conn *link.Conn) {
	var remoteID = conn.Link().RemoteIdentity()
	var peer = mod.node.Network().Peers().Find(remoteID)

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

			if err := mod.move(conn, preferred); err != nil {
				mod.log.Error("error moving connection: %s", err)
			} else {
				if current.Conns().Count() == 0 {
					current.Idle().SetTimeout(5 * time.Minute)
				}
			}
		}
	}
}

func (mod *Module) move(movable *link.Conn, targetLink *link.Link) error {
	var srcNet = movable.Link().Network()
	var dstNet = targetLink.Network()
	var query = movable.Query()

	mod.log.Logv(1, "moving %s from %s to %s", query, srcNet, dstNet)

	moveID, err := mod.pick(movable)
	if err != nil {
		return err
	}

	err = mod.drop(targetLink, movable, moveID)
	if err == nil {
		mod.log.Info("moved %s from %s to %s", query, srcNet, dstNet)
	}
	return err
}

func (mod *Module) pick(movable *link.Conn) (int, error) {
	// start transfer on the source link
	server, err := movable.Link().Query(context.Background(), portPick)
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

func (mod *Module) drop(targetLink *link.Link, movable *link.Conn, moveID int) error {
	// allocate a new port on the destination link
	var newReader = link.NewPortReader()
	newLocalPort, err := targetLink.Mux().BindAny(newReader)
	if err != nil {
		return err
	}

	// set connection to fall back to the new input stream
	movable.SetFallbackReader(newReader)

	// start the query
	query, err := targetLink.Query(context.Background(), portDrop)
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

func (mod *Module) unusedMoveID() int {
	for i := 0; i < 255; i++ {
		if _, ok := mod.moves[i]; !ok {
			return i
		}
	}
	return -1
}

func (mod *Module) allocMove(conn *link.Conn) int {
	id := mod.unusedMoveID()
	if id != -1 {
		mod.moves[id] = conn
	}
	return id
}
