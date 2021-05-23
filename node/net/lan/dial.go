package lan

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
	goNet "net"
)

func (drv *driver) Dial(_ context.Context, addr net.Addr) (net.Conn, error) {

	// Check if IP is in a local network
	if !isAddrLocalNetwork(addr) {
		log.Println("not lan", addr.String())
	}

	tcpConn, err := goNet.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return net.WrapConn(tcpConn, true), nil
}
