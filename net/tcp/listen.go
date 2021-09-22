package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/net/ip"
	"github.com/cryptopunkscc/astrald/net/mon"
	"log"
	go_net "net"
	"strconv"
)

const tcpPort = 1791

func (drv *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	output := make(chan net.Conn)

	go func() {
		defer close(output)

		for name := range mon.Interfaces(ctx) {
			go func(name string) {
				for conn := range listenInterface(ctx, name) {
					output <- conn
				}
			}(name)
		}
	}()

	return output, nil
}

func listenInterface(ctx context.Context, ifaceName string) <-chan net.Conn {
	output := make(chan net.Conn)

	go func() {
		defer close(output)

		ifaceCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		for addr := range mon.Addrs(ifaceCtx, ifaceName) {
			go func(addr string) {
				for conn := range listenAddr(ifaceCtx, addr) {
					output <- conn
				}
			}(addr)
		}
	}()

	return output
}

func listenAddr(ctx context.Context, addr string) <-chan net.Conn {
	output := make(chan net.Conn)

	go func() {
		defer close(output)

		ip, _ := ip.SplitIPMask(addr)
		hostPort := go_net.JoinHostPort(ip, strconv.Itoa(tcpPort))

		listener, err := go_net.Listen("tcp", hostPort)
		if err != nil {
			return
		}

		log.Println("listen tcp", hostPort)

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			output <- net.WrapConn(conn, false)
		}

		log.Println("closed tcp", hostPort)
	}()

	return output
}
