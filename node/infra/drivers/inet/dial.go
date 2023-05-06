package inet

import (
	"context"
	net "github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	_net "net"
	"time"
)

const dialTimeout = 10 * time.Second
const dialTimeoutPrivate = 3 * time.Second

var dialConfig = _net.Dialer{Timeout: dialTimeout}

var _ infra.Dialer = &Driver{}

func (drv *Driver) Dial(ctx context.Context, endpoint net.Endpoint) (net.Conn, error) {
	var config = dialConfig

	endpoint, err := drv.Unpack(endpoint.Network(), endpoint.Pack())
	if err != nil {
		return nil, err
	}

	inetEndpoint := endpoint.(Endpoint)

	// for LAN dials we can use a shorter timeout
	if inetEndpoint.IsPrivate() {
		config.Timeout = dialTimeoutPrivate
	}

	tcpConn, err := config.DialContext(ctx, "tcp", endpoint.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}

func (drv *Driver) DialFrom(ctx context.Context, addr Endpoint, from Endpoint) (net.Conn, error) {
	var err error
	var config = dialConfig

	config.LocalAddr, err = _net.ResolveTCPAddr("tcp", from.String())
	if err != nil {
		return nil, err
	}

	tcpConn, err := config.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return newConn(tcpConn, true), nil
}
