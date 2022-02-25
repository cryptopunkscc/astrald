package roam

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"log"
	"sync"
	"time"
)

const (
	portPick   = "roam.pick"
	portDrop   = "roam.drop"
	ModuleName = "roam"
	logTag     = "(roam)"
)

type Roam struct {
	node  *_node.Node
	moves map[int]*link.Conn
}

func (roam *Roam) Run(ctx context.Context, node *_node.Node) error {
	roam.moves = make(map[int]*link.Conn)
	roam.node = node

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
	for event := range roam.node.Subscribe(ctx.Done()) {
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

		query := query.Accept()

		// read remote stream id of the connection to pick
		remoteStreamID, _ := cslq.ReadUint16(query)

		// find the connection
		for c := range query.Link().Conns() {
			if c.OutputStream().ID() == int(remoteStreamID) {
				// allocate a new move for the connection
				moveID := roam.allocMove(c)
				if moveID != -1 {
					cslq.Write(query, uint8(moveID))
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

		query := query.Accept()

		// read moveID and remote output ID
		moveID, _ := cslq.ReadUint8(query)
		newOutputID, _ := cslq.ReadUint16(query)

		conn, found := roam.moves[int(moveID)]
		if !found {
			query.Close()
			continue
		}

		// allocate a new input stream and write its id
		newInputStream, _ := query.Link().AllocInputStream()
		conn.SetFallbackInputStream(newInputStream)
		cslq.Write(query, uint16(newInputStream.ID()))

		// replace the output stream and finalize the move
		newOutputStream := query.Link().OutputStream(int(newOutputID))
		conn.ReplaceOutputStream(newOutputStream)
		conn.Attach(query.Link())

		query.Close()
	}
}

func (roam *Roam) optimizeConn(conn *link.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var remoteID = conn.Link().RemoteIdentity()
	var peer = roam.node.Peers.Hold(ctx, remoteID)

	for {
		select {
		case <-conn.Wait():
			return
		case <-time.After(time.Second):
			best := link.Select(peer.Links(), link.LowestRoundTrip)

			if conn.Link() != best {
				if err := roam.move(conn, best); err != nil {
					log.Println(logTag, "move error:", err)
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
	cslq.Write(init, uint16(conn.InputStream().ID()))

	// read transfer id
	tid, err := cslq.ReadUint8(init)
	if err != nil {
		return -1, err
	}

	return int(tid), nil
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
		newInputStream.Close()
		return err
	}
	defer query.Close()

	// write move id and new input stream id
	cslq.Write(query, uint8(moveID))
	cslq.Write(query, uint16(newInputStream.ID()))

	// preapre the new output stream
	newOutputStreamID, _ := cslq.ReadUint16(query)
	if newOutputStreamID == 0 {
		newInputStream.Close()
		return errors.New("received invalid id")
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

func (Roam) String() string {
	return ModuleName
}
