//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"golang.org/x/sys/unix"
)

var _ infra.Dialer = &Driver{}

func (drv *Driver) Dial(ctx context.Context, endpoint net.Endpoint) (conn net.Conn, err error) {
	endpoint, err = drv.Unpack(endpoint.Network(), endpoint.Pack())
	if err != nil {
		return nil, err
	}

	btEndpoint := endpoint.(Endpoint)

	var fd int
	var done = make(chan struct{})
	defer close(done)

	fd, err = unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return
	}

	go func() {
		select {
		case <-ctx.Done():
			unix.Close(fd)
		case <-done:
		}
	}()

	var mac [6]byte
	copy(mac[:], btEndpoint.Pack())

	err = unix.Connect(fd, &unix.SockaddrRFCOMM{
		Channel: 1,
		Addr:    mac,
	})

	if err != nil {
		unix.Close(fd)
		return
	}

	conn = &Conn{
		nfd:            fd,
		outbount:       true,
		localEndpoint:  Endpoint{},
		remoteEndpoint: btEndpoint,
	}

	return
}
