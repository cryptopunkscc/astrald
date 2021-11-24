package info

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/network/peer"
	"io"
)

const serviceHandle = "info"

func info(ctx context.Context, node *node.Node) error {
	port, err := node.Ports.Register(serviceHandle)
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
			conn.Write(node.Network.Info(false).Pack())
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
					queryInfo(node, event.Peer)
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

func queryInfo(node *node.Node, peer *peer.Peer) {
	// update peer info
	conn, err := peer.Query(serviceHandle)
	if err != nil {
		return
	}
	packed, err := io.ReadAll(conn)
	if err == nil {
		node.Network.Graph.AddPacked(packed)
	}
}

func init() {
	_ = node.RegisterService(serviceHandle, info)
}
