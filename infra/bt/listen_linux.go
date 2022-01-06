//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt/bluez"
	"github.com/cryptopunkscc/astrald/log"
	"golang.org/x/sys/unix"
	"syscall"
)

var _ infra.Listener = &Bluetooth{}

func (bt Bluetooth) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	output := make(chan infra.Conn)

	b, err := bluez.New()
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(output)

		err = b.RegisterProfile(bluez.UUID_SPP)
		if err != nil {
			debugf("(%s) RegisterProfile error: %s\n", NetworkName, err.Error())
			return
		}

		btSocket, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
		if err != nil {
			debugf("(%s) unix.Socket error: %s\n", NetworkName, err.Error())
			return
		}

		addr := &unix.SockaddrRFCOMM{
			Addr:    [6]byte{0, 0, 0, 0, 0, 0},
			Channel: 1,
		}

		err = unix.Bind(btSocket, addr)
		if err != nil {
			debugf("(%s) unix.Bind error: %s\n", NetworkName, err.Error())
			return
		}

		err = unix.Listen(btSocket, 1)
		if err != nil {
			debugf("(%s) unix.Listen error: %s\n", NetworkName, err.Error())
			return
		}

		var addrs []string
		for _, as := range bt.Addresses() {
			addrs = append(addrs, as.String())
		}
		log.Printf("(%s) listen %s\n", NetworkName, addrs)

		go func() {
			<-ctx.Done()
			unix.Close(btSocket)
		}()

		defer unix.Close(btSocket)

		for {
			nfd, sa, err := unix.Accept(btSocket)
			if err != nil {
				debugf("(%s) unix.Accept error: %s\n", NetworkName, err.Error())
				return
			}

			remoteAddr, _ := Unpack(sa.(*unix.SockaddrRFCOMM).Addr[:])

			log.Printf("(%s) accepted %s\n",
				NetworkName,
				remoteAddr.String())

			output <- Conn{
				nfd:        nfd,
				outbount:   false,
				localAddr:  Addr{},
				remoteAddr: remoteAddr,
			}
		}

	}()

	return output, nil
}
