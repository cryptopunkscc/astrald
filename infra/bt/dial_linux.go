//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/sys/unix"
)

func (bt Bluetooth) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	_addr, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return nil, err
	}

	var mac [6]byte
	copy(mac[:], _addr.Pack())

	err = unix.Connect(fd, &unix.SockaddrRFCOMM{
		Channel: 1,
		Addr:    mac,
	})

	if err != nil {
		return nil, err
	}

	return &Conn{
		nfd:        fd,
		outbount:   true,
		localAddr:  Addr{},
		remoteAddr: _addr,
	}, err
}
