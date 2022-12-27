//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/sys/unix"
)

func (bt *Bluetooth) Dial(ctx context.Context, addr Addr) (conn infra.Conn, err error) {
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
	copy(mac[:], addr.Pack())

	err = unix.Connect(fd, &unix.SockaddrRFCOMM{
		Channel: 1,
		Addr:    mac,
	})

	if err != nil {
		unix.Close(fd)
		return
	}

	conn = &Conn{
		nfd:        fd,
		outbount:   true,
		localAddr:  Addr{},
		remoteAddr: addr,
	}

	return
}
