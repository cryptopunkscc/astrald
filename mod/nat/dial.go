package nat

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	"github.com/cryptopunkscc/astrald/node/network"
	"time"
)

func (mod *Module) makeLink(ctx context.Context, remoteAddr inet.Endpoint, remoteID id.Identity) (conn net.SecureConn, err error) {
	log.Logv(1, "dial %s", remoteAddr)

	drv, ok := infra.GetDriver[*inet.Driver](mod.node.Infra(), inet.DriverName)
	if !ok {
		return nil, errors.New("inet unsupported")
	}

	startTime := time.Now()
	dialCtx, _ := context.WithTimeout(ctx, dialTimeout)
	newConn, err := drv.DialFrom(dialCtx, remoteAddr, mod.mapping.intAddr)
	dialTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	if err != nil {
		log.Errorv(1, "dial %s after %.3fms: %s", remoteAddr, dialTime, err)
		return nil, err
	}

	log.Info("dial %s success after %.3fms!", remoteAddr, dialTime)

	hsCtx, _ := context.WithTimeout(ctx, network.HandshakeTimeout)
	if remoteID.IsZero() {
		conn, err = auth.HandshakeInbound(hsCtx, inboundConn{newConn}, mod.node.Identity())
	} else {
		conn, err = auth.HandshakeOutbound(hsCtx, newConn, remoteID, mod.node.Identity())
	}

	if err != nil {
		log.Error("handshake error: %s", err)
		newConn.Close()
		return nil, err
	}

	log.Info("successfully traversed with %s via %s", remoteID, remoteAddr)

	return
}
