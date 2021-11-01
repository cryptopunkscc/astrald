package route

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network"
	"io"
)

func route(ctx context.Context, node *node.Node) error {
	port, err := node.Ports.Register("route")
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		port.Close()
	}()

	go func() {
		for req := range port.Requests() {
			conn := req.Accept()
			conn.Write(node.Network.Route(false).Pack())
			conn.Close()
		}
	}()

	go func() {
		future := node.FutureEvent()
		for {
			select {
			case <-future.Wait():
				e := future.Event()
				if e.Event() == network.EventPeerLinked {
					event := e.(network.Event)
					queryRoute(node, event.Peer)
				}
				future = future.Next()
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	port.Close()

	return nil
}

func queryRoute(node *node.Node, peer *network.Peer) {
	// update peer's routes
	conn, err := peer.Query("route")
	if err != nil {
		return
	}
	packedRoute, err := io.ReadAll(conn)
	if err == nil {
		node.Network.Router.AddPacked(packedRoute)
	}
}

func init() {
	_ = node.RegisterService("route", route)
}
