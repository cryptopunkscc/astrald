package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"log"
)

func (c *Server) handleRegister(ctx context.Context) error {
	var (
		portName string
		target   string
	)

	c.Decode("[c]c [c]c", &portName, &target)

	port, err := c.node.Ports.RegisterContext(ctx, portName)
	if err != nil {
		return c.closeWithError(err)
	}

	c.Encode("c", proto.ResponseOK)

	// cancel the context when the control connection closes
	go func() {
		for {
			var buf [1024]byte
			_, err := c.conn.Read(buf[:])
			if err != nil {
				c.cancel()
				return
			}
		}
	}()

	for query := range port.Queries() {
		remoteID := c.node.Identity()
		if !query.IsLocal() {
			remoteID = query.Link().RemoteIdentity()
		}

		log.Printf("(apphost) [%s] <- %s\n", c.node.Contacts.DisplayName(remoteID), query.Query())

		appConn, err := ipc.Dial(target)
		if err != nil {
			query.Reject()
			continue
		}

		appStream := cslq.NewEndec(appConn)

		if err := appStream.Encode("v [c]c", remoteID, query.Query()); err != nil {
			log.Println("[apphost] encode error:", err)
			appConn.Close()
			continue
		}

		var result int

		if err := appStream.Decode("c", &result); err != nil {
			log.Println("[apphost] decode error:", err)
			appConn.Close()
			continue
		}

		if result != proto.ResponseOK {
			query.Reject()
			continue
		}

		go join(ctx, appConn, query.Accept())
	}

	c.conn.Close()

	return nil
}
