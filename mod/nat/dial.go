package nat

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/node/peers"
	"log"
	"time"
)

func (mod *Module) makeLink(ctx context.Context, remoteAddr inet.Addr, remoteID id.Identity) (conn auth.Conn, err error) {
	log.Printf("[nat] dial %s\n", remoteAddr.String())

	startTime := time.Now()
	dialCtx, _ := context.WithTimeout(ctx, dialTimeout)
	newConn, err := mod.node.Infra.Inet().DialFrom(dialCtx, remoteAddr, mod.mapping.intAddr)
	dialTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	if err != nil {
		log.Printf("[nat] dial %s error after %.3fms: %s\n", remoteAddr.String(), dialTime, err)
		return nil, err
	}

	log.Printf("[nat] dial %s success after %.3fms!\n", remoteAddr.String(), dialTime)

	var authed auth.Conn

	hsCtx, _ := context.WithTimeout(ctx, peers.HandshakeTimeout)
	if remoteID.IsZero() {
		authed, err = auth.HandshakeInbound(hsCtx, inboundConn{newConn}, mod.node.Identity())
	} else {
		authed, err = auth.HandshakeOutbound(hsCtx, newConn, remoteID, mod.node.Identity())
	}

	if err != nil {
		log.Printf("[nat] handshake error: %s\n", err.Error())
		newConn.Close()
		return nil, err
	}

	log.Printf("[nat] successfully traversed via %s\n", remoteAddr.String())

	return authed, nil
}
