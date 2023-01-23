//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt/bluez"
	"golang.org/x/sys/unix"
	"syscall"
)

var _ infra.Listener = &Bluetooth{}

func (bt *Bluetooth) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	output := make(chan infra.Conn)

	b, err := bluez.New()
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(output)

		err = b.RegisterProfile(bluez.UUID_SPP)
		if err != nil {
			log.Errorv(1, "RegisterProfile: %s", err.Error())
			return
		}

		btSocket, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
		if err != nil {
			log.Errorv(1, "unix.Socket: %s", err.Error())
			return
		}

		addr := &unix.SockaddrRFCOMM{
			Addr:    [6]byte{0, 0, 0, 0, 0, 0},
			Channel: 1,
		}

		err = unix.Bind(btSocket, addr)
		if err != nil {
			log.Errorv(1, "unix.Bind: %s", err.Error())
			return
		}

		err = unix.Listen(btSocket, 1)
		if err != nil {
			log.Errorv(1, " unix.Listen: %s", err.Error())
			return
		}

		var addrs []string
		for _, as := range bt.Addresses() {
			addrs = append(addrs, as.String())
		}
		log.Logv(1, "listen %s", addrs)

		go func() {
			<-ctx.Done()
			unix.Close(btSocket)
		}()

		defer unix.Close(btSocket)

		for {
			nfd, sa, err := unix.Accept(btSocket)
			if err != nil {
				log.Errorv(1, "unix.Accept: %s", err.Error())
				return
			}

			remoteAddr, _ := Unpack(sa.(*unix.SockaddrRFCOMM).Addr[:])

			log.Log("accepted %s %s",
				log.Em(NetworkName),
				remoteAddr)

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
