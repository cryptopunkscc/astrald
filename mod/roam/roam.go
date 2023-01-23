package roam

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	_log "github.com/cryptopunkscc/astrald/log"
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
	node  *node.Node
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
	for event := range mod.node.Subscribe(ctx) {
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
		if q == portDrop || q == portPick || q[0] == '.' {
			continue
		}

		go mod.optimizeConn(e.Conn)
	}
}

func (mod *Module) servePick(ctx context.Context) error {
	port, err := mod.node.Ports.RegisterContext(ctx, portPick)
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

		var remoteStreamID uint16

		// read remote stream id of the connection to pick
		cslq.Decode(query, "s", &remoteStreamID)

		// find the connection
		for c := range query.Link().Conns() {
			if c.OutputStream().ID() == int(remoteStreamID) {
				// allocate a new move for the connection
				moveID := mod.allocMove(c)
				if moveID != -1 {
					cslq.Encode(query, "c", moveID)
				}
				break
			}
		}
		query.Close()
	}

	return nil
}

func (mod *Module) serveDrop(ctx context.Context) {
	port, err := mod.node.Ports.RegisterContext(ctx, portDrop)
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

		var moveID, newOutputID int

		cslq.Decode(conn, "cs", &moveID, &newOutputID)

		target, found := mod.moves[moveID]
		if !found {
			conn.Close()
			continue
		}

		// allocate a new input stream and write its id
		newInputStream, _ := conn.Link().AllocInputStream()
		target.SetFallbackInputStream(newInputStream)

		cslq.Encode(conn, "s", newInputStream.ID())

		// replace the output stream and finalize the move
		newOutputStream := conn.Link().OutputStream(int(newOutputID))
		target.ReplaceOutputStream(newOutputStream)
		target.Attach(conn.Link())

		conn.Close()
	}
}

func (mod *Module) optimizeConn(conn *link.Conn) {
	var remoteID = conn.Link().RemoteIdentity()
	var peer = mod.node.Peers.Pool.Peer(remoteID)

	for {
		select {
		case <-conn.Wait():
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

			// only move to a more preferred network (avoid unnecessary moves due to ping jitter)
			if scoreNet(preferred.Network()) <= scoreNet(current.Network()) {
				continue
			}

			if err := mod.move(conn, preferred); err != nil {
				log.Error("move: %s", err)
			} else {
				if current.ConnCount() == 0 {
					current.SetIdleTimeout(time.Minute)
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
	cslq.Encode(init, "s", conn.InputStream().ID())

	// read transfer id
	var tid int
	err = cslq.Decode(init, "c", &tid)
	if err != nil {
		return -1, err
	}

	return tid, nil
}

func (mod *Module) drop(dest *link.Link, conn *link.Conn, moveID int) error {
	// allocate input stream on the destination link
	newInputStream, err := dest.AllocInputStream()
	if err != nil {
		return err
	}

	// set connection to fall back to the new input stream
	conn.SetFallbackInputStream(newInputStream)

	// start the query
	query, err := dest.Query(context.Background(), portDrop)
	if err != nil {
		newInputStream.Discard()
		return err
	}
	defer query.Close()

	// write move id and new input stream id
	cslq.Encode(query, "cs", moveID, newInputStream.ID())

	// preapre the new output stream
	var newOutputStreamID uint16
	err = cslq.Decode(query, "s", &newOutputStreamID)
	if err != nil {
		newInputStream.Discard()
		return fmt.Errorf("failed to read output stream: %w", err)
	}
	newOutputStream := dest.OutputStream(int(newOutputStreamID))

	// finalize the move
	conn.ReplaceOutputStream(newOutputStream)
	conn.Attach(dest)

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

func scoreNet(net string) int {
	switch net {
	case tor.NetworkName:
		return 10
	case bt.NetworkName:
		return 20
	case gw.NetworkName:
		return 30
	case inet.NetworkName:
		return 40
	}
	return 0
}
