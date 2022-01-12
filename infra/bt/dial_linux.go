//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/sys/unix"
	"log"
)

func (bt Bluetooth) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	log.Println("BT Connecting to", addr.String()) // TODO: for debugging testing purpose, remove later

	_addr, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		log.Println("Unix socket error", err) // TODO: for debugging testing purpose, remove later
		return nil, err
	}

	var mac [6]byte
	copy(mac[:], _addr.Pack())

	err = unix.Connect(fd, &unix.SockaddrRFCOMM{
		Channel: 1,
		Addr:    mac,
	})

	if err != nil {
		log.Println("Unix socket connect error", err) // TODO: for debugging testing purpose, remove later
		return nil, err
	}

	return &Conn{
		nfd:        fd,
		outbount:   true,
		localAddr:  Addr{},
		remoteAddr: _addr,
	}, err
}
