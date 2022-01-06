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
			log.Printf("(%s) RegisterProfile error: %s\n", NetworkName, err.Error())
			return
		}

		btSocket, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
		if err != nil {
			log.Printf("(%s) unix.Socket error: %s\n", NetworkName, err.Error())
			return
		}

		addr := &unix.SockaddrRFCOMM{
			Addr:    [6]byte{0, 0, 0, 0, 0, 0},
			Channel: 1,
		}

		err = unix.Bind(btSocket, addr)
		if err != nil {
			log.Printf("(%s) unix.Bind error: %s\n", NetworkName, err.Error())
			return
		}

		err = unix.Listen(btSocket, 1)
		if err != nil {
			log.Printf("(%s) unix.Listen error: %s\n", NetworkName, err.Error())
			return
		}

		defer unix.Close(btSocket)

		var addrs []string

		for _, as := range bt.Addresses() {
			addrs = append(addrs, as.String())
		}

		log.Printf("(%s) listen %s\n", NetworkName, addrs)

		for {
			nfd, sa, err := unix.Accept(btSocket)
			if err != nil {
				log.Printf("(%s) unix.Bind error: %s\n", NetworkName, err.Error())
				return
			}

			remoteAddr, _ := Unpack(sa.(*unix.SockaddrRFCOMM).Addr[:])

			log.Printf("(%s) accept: %s\n",
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
