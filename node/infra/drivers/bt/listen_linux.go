//go:build linux
// +build linux

package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/bt/bluez"
	"golang.org/x/sys/unix"
	"syscall"
)

var _ infra.Listener = &Driver{}

func (drv *Driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	output := make(chan net.Conn)

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

		var endpoints []string
		for _, endpoint := range drv.Endpoints() {
			log.Logv(1, "listen %s", endpoint)
			endpoints = append(endpoints, endpoint.String())
		}

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

			remoteEndpoint, _ := Unpack(sa.(*unix.SockaddrRFCOMM).Addr[:])

			log.Log("accepted %s %s",
				DriverName,
				remoteEndpoint)

			output <- &Conn{
				nfd:            nfd,
				outbount:       false,
				localEndpoint:  Endpoint{},
				remoteEndpoint: remoteEndpoint,
			}
		}

	}()

	return output, nil
}
