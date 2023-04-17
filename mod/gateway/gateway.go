package gateway

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra/gw"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/streams"
)

const queryConnect = "connect"

type Gateway struct {
	node node.Node
}

var log = _log.Tag(ModuleName)

func (mod *Gateway) Run(ctx context.Context) error {
	port, err := mod.node.Services().RegisterContext(ctx, gw.PortName)
	if err != nil {
		return err
	}

	for req := range port.Queries() {
		conn, err := req.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := mod.handleConn(ctx, conn); err != nil {
				cslq.Encode(conn, "c", false)
				log.Error("serve error: %s", err)
			}
			defer conn.Close()
		}()
	}

	return nil
}

func (mod *Gateway) handleConn(ctx context.Context, conn *hub.Conn) error {
	c := cslq.NewEndec(conn)

	var cookie string

	err := c.Decode("[c]c", &cookie)
	if err != nil {
		return err
	}

	nodeID, err := id.ParsePublicKeyHex(cookie)
	if err != nil {
		return err
	}

	peer := mod.node.Network().Peers().Find(nodeID)
	if peer == nil {
		return errors.New("node unavailable")
	}

	lnk := peer.PreferredLink()
	if lnk == nil {
		return errors.New("node unavailable")
	}

	out, err := lnk.Query(ctx, queryConnect)
	if err != nil {
		conn.Close()
		return err
	}

	c.Encode("c", true)

	l, r, err := streams.Join(conn, out)

	log.Logv(1, "conn for %s done (bytes read %d written %d)", peer.Identity(), l, r)

	return err
}
