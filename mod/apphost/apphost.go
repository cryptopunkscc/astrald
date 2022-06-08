package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/proto/apphost"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"log"
	"net"
	"sync"
)

var _ apphost.AppHost = &AppHost{}

type AppHost struct {
	ports *PortManager
	mu    sync.Mutex

	node *node.Node
	conn net.Conn
}

func (host *AppHost) Register(portName string, target string) error {
	if host.ports.GetPort(portName) != nil {
		return hub.ErrAlreadyRegistered
	}

	port, err := host.node.Ports.Register(portName)
	if err != nil {
		return err
	}

	if err := host.ports.AddPort(port, target); err != nil {
		return err
	}

	go func() {
		for q := range port.Queries() {
			q := q
			go func() {
				remoteID := host.node.Identity()
				if !q.IsLocal() {
					remoteID = q.Link().RemoteIdentity()
				}

				conn, err := ipc.Dial(target)
				if err != nil {
					q.Reject()
					port.Close()
					log.Printf("[apphost] target %s unreachable, closing port %s\n", target, portName)
					return
				}

				defer conn.Close()

				stream := cslq.NewEndec(conn)

				if err := stream.Encode("v [c]c", remoteID, q.Query()); err != nil {
					log.Println("[apphost] encode error:", err)
					return
				}

				var errorCode int

				if err := stream.Decode("c", &errorCode); err != nil {
					log.Println("[apphost] decode error:", err)
					return
				}

				if errorCode != 0 {
					q.Reject()
					return
				}

				streams.Join(conn, q.Accept())
			}()
		}
	}()

	return nil
}

func (host *AppHost) Query(identity id.Identity, query string) (io.ReadWriteCloser, error) {
	return host.node.Query(context.Background(), identity, query)
}

func (host *AppHost) Resolve(s string) (id.Identity, error) {
	if identity, err := id.ParsePublicKeyHex(s); err == nil {
		return identity, nil
	}

	if s == "localnode" {
		return host.node.Identity(), nil
	}

	return host.node.Contacts.ResolveIdentity(s)
}

func (host *AppHost) NodeInfo(identity id.Identity) (apphost.NodeInfo, error) {
	return apphost.NodeInfo{
		Identity: identity,
		Name:     host.node.Contacts.DisplayName(identity),
	}, nil
}
