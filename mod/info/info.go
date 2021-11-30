package info

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/network/contacts"
	"github.com/cryptopunkscc/astrald/node/network/peer"
	"io"
	"log"
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
					if err := queryInfo(node, event.Peer); err != nil {
						if !errors.Is(err, link.ErrRejected) {
							log.Println("[info] error fetching info:", err)
						}
					}
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

func queryInfo(node *node.Node, peer *peer.Peer) error {
	// update peer info
	conn, err := peer.Query(serviceHandle)
	if err != nil {
		return err
	}
	packed, err := io.ReadAll(conn)
	if err != nil {
		return err
	}

	info, err := contacts.Unpack(packed)
	if err != nil {
		return err
	}

	node.Network.Contacts.AddInfo(info)

	return nil
}

func init() {
	_ = node.RegisterService(serviceHandle, info)
}
