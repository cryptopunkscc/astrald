package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"strconv"
)

const defaultPort = 1791

func listenInterface(ctx context.Context, ifaceName string) <-chan infra.Conn {
	output := make(chan infra.Conn)

	go func() {
		defer close(output)

		ifaceCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		for addr := range ip.Addrs(ifaceCtx, ifaceName) {
			go func(addr string) {
				for conn := range listenAddr(ifaceCtx, addr) {
					output <- conn
				}
			}(addr)
		}
	}()

	return output
}

func listenAddr(ctx context.Context, addr string) <-chan infra.Conn {
	output := make(chan infra.Conn)

	go func() {
		defer close(output)

		ip, _ := ip.SplitIPMask(addr)
		hostPort := net.JoinHostPort(ip, strconv.Itoa(defaultPort))

		listener, err := net.Listen("tcp", hostPort)
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
			output <- newConn(conn, false)
		}

		log.Println("closed tcp", hostPort)
	}()

	return output
}
