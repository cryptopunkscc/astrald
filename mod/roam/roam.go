package roam

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
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
}

var log = _log.Tag(ModuleName)

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
	port, err := mod.node.Services().RegisterContext(ctx, portPick)
	if err != nil {
		return err
	}

	for query := range port.Queries() {
		// skip local queries
		if query.IsLocal() {
			query.Reject()
			continue
		}

		query, err := query.Accept()
		if err != nil {
			continue
		}

		var remotePort int

		// read remote port of the connection to pick
		cslq.Decode(query, "s", &remotePort)

		// find the connection
		c := query.Link().Conns().FindByRemotePort(remotePort)
		if c == nil {
			query.Close()
			continue
		}

		// allocate a new move for the connection
		moveID := mod.allocMove(c)
		if moveID != -1 {
			cslq.Encode(query, "c", moveID)
		}

		query.Close()
	}

	return nil
}

func (mod *Module) serveDrop(ctx context.Context) {
	port, err := mod.node.Services().RegisterContext(ctx, portDrop)
	if err != nil {
		return
	}

	for query := range port.Queries() {
		// skip local queries
		if query.IsLocal() {
			query.Reject()
			continue
		}

		conn, _ := query.Accept()

		var moveID, newRemotePort int

		cslq.Decode(conn, "cs", &moveID, &newRemotePort)

		target, found := mod.moves[moveID]
		if !found {
			conn.Close()
			continue
		}

		// allocate a new input stream and write its id
		var newReader = mux.NewFrameReader()
		newLocalPort, err := conn.Link().Mux().BindAny(newReader)
		if err != nil {
			conn.Link().Mux().Unbind(newLocalPort)
			conn.Close()
			continue
		}

		target.SetFallbackReader(newReader)

		cslq.Encode(conn, "s", newLocalPort)

		// replace the output stream and finalize the move
		newWriter := mux.NewFrameWriter(conn.Link().Mux(), newRemotePort)
		target.SetWriter(newWriter)
		target.Attach(conn.Link())

		conn.Close()
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
				log.Error("move: %s", err)
			} else {
				if current.Conns().Count() == 0 {
					current.Idle().SetTimeout(time.Minute)
				}
			}
		}
	}
}

func (mod *Module) move(conn *link.Conn, dest *link.Link) error {
	log.Log("moving %s from %s to %s", conn.Query(), conn.Link().Network(), dest.Network())

	moveID, err := mod.init(conn)
	if err != nil {
		return err
	}

	return mod.drop(dest, conn, moveID)
}

func (mod *Module) init(conn *link.Conn) (int, error) {
	// start transfer on the source link
	init, err := conn.Link().Query(context.Background(), portPick)
	if err != nil {
		log.Error("init rejected (%s)", err)
		return -1, err
	}
	defer init.Close()

	// write input stream id of the connection to be migrated
	cslq.Encode(init, "s", conn.LocalPort())

	// read transfer id
	var tid int
	err = cslq.Decode(init, "c", &tid)
	if err != nil {
		return -1, err
	}

	return tid, nil
}

func (mod *Module) drop(target *link.Link, conn *link.Conn, moveID int) error {
	// allocate a new port on the destination link
	var newReader = mux.NewFrameReader()
	newLocalPort, err := target.Mux().BindAny(newReader)
	if err != nil {
		return err
	}

	// set connection to fall back to the new input stream
	conn.SetFallbackReader(newReader)

	// start the query
	query, err := target.Query(context.Background(), portDrop)
	if err != nil {
		target.Mux().Unbind(newLocalPort)
		return err
	}
	defer query.Close()

	// write move id and new input stream id
	cslq.Encode(query, "cs", moveID, newLocalPort)

	// preapre the new output stream
	var newRemotePort int
	err = cslq.Decode(query, "s", &newRemotePort)
	if err != nil {
		target.Mux().Unbind(newLocalPort)
		return fmt.Errorf("read error: %w", err)
	}
	newWriter := mux.NewFrameWriter(target.Mux(), newRemotePort)

	// finalize the move
	conn.SetWriter(newWriter)
	conn.Attach(target)

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
