package nat

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/node/peers"
	"time"
)

func (mod *Module) makeLink(ctx context.Context, remoteAddr inet.Addr, remoteID id.Identity) (conn auth.Conn, err error) {
	log.Logv(1, "dial %s", remoteAddr)

	startTime := time.Now()
	dialCtx, _ := context.WithTimeout(ctx, dialTimeout)
	newConn, err := mod.node.Infra().Inet().DialFrom(dialCtx, remoteAddr, mod.mapping.intAddr)
	dialTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	if err != nil {
		log.Errorv(1, "dial %s after %.3fms: %s", remoteAddr, dialTime, err)
		return nil, err
	}

	log.Info("dial %s success after %.3fms!", remoteAddr, dialTime)

	var authed auth.Conn

	hsCtx, _ := context.WithTimeout(ctx, peers.HandshakeTimeout)
	if remoteID.IsZero() {
		authed, err = auth.HandshakeInbound(hsCtx, inboundConn{newConn}, mod.node.Identity())
	} else {
		authed, err = auth.HandshakeOutbound(hsCtx, newConn, remoteID, mod.node.Identity())
	}

	if err != nil {
		log.Error("handshake error: %s", err)
		newConn.Close()
		return nil, err
	}

	log.Info("successfully traversed with %s via %s", remoteID, remoteAddr)

	return authed, nil
}
