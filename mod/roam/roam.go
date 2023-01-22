package roam

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"log"
	"sync"
	"time"
)

const (
	portPick = "roam.pick"
	portDrop = "roam.drop"
	logTag   = "(roam)"
)

type Roam struct {
	node  *node.Node
	moves map[int]*link.Conn
}

func (roam *Roam) Run(ctx context.Context) error {
	roam.moves = make(map[int]*link.Conn)

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		roam.servePick(ctx)
		wg.Done()
	}()

	go func() {
		roam.serveDrop(ctx)
		wg.Done()
	}()

	go func() {
		roam.monitorConnections(ctx)
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (roam *Roam) monitorConnections(ctx context.Context) {
	for event := range roam.node.Subscribe(ctx) {
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

		go roam.optimizeConn(e.Conn)
	}
}

func (roam *Roam) servePick(ctx context.Context) error {
	port, err := roam.node.Ports.RegisterContext(ctx, portPick)
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
				moveID := roam.allocMove(c)
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

func (roam *Roam) serveDrop(ctx context.Context) {
	port, err := roam.node.Ports.RegisterContext(ctx, portDrop)
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

		target, found := roam.moves[moveID]
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

func (roam *Roam) optimizeConn(conn *link.Conn) {
	var remoteID = conn.Link().RemoteIdentity()
	var peer = roam.node.Peers.Pool.Peer(remoteID)

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

			if err := roam.move(conn, preferred); err != nil {
				log.Println(logTag, "move error:", err)
			} else {
				if current.ConnCount() == 0 {
					current.SetIdleTimeout(time.Minute)
				}
			}
		}
	}
}

func (roam *Roam) move(conn *link.Conn, dest *link.Link) error {
	log.Println(logTag, "move", conn.Query(), "from", conn.Link().Network(), "to", dest.Network())

	moveID, err := roam.init(conn)
	if err != nil {
		return err
	}

	return roam.drop(dest, conn, moveID)
}

func (roam *Roam) init(conn *link.Conn) (int, error) {
	// start transfer on the source link
	init, err := conn.Link().Query(context.Background(), portPick)
	if err != nil {
		log.Println("(transfer) init rejected")
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

func (roam *Roam) drop(dest *link.Link, conn *link.Conn, moveID int) error {
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

func (roam *Roam) unusedMoveID() int {
	for i := 0; i < 255; i++ {
		if _, ok := roam.moves[i]; !ok {
			return i
		}
	}
	return -1
}

func (roam *Roam) allocMove(conn *link.Conn) int {
	id := roam.unusedMoveID()
	if id != -1 {
		roam.moves[id] = conn
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
