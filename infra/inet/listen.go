package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"strconv"
	"strings"
)

func (inet Inet) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	if inet.separateListen {
		return inet.listenSeparately(ctx)
	}
	return inet.listenCombined(ctx)
}

func (inet Inet) listenCombined(ctx context.Context) (<-chan infra.Conn, error) {
	output := make(chan infra.Conn)

	go func() {
		defer close(output)

		hostPort := "0.0.0.0:" + strconv.Itoa(int(inet.listenPort))

		l, err := net.Listen("tcp", hostPort)
		if err != nil {
			return
		}

		log.Println("listen tcp", hostPort)

		go func() {
			<-ctx.Done()
			l.Close()
		}()

		for {
			conn, err := l.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("accept error:", err)
				}
				return
			}

			output <- newConn(conn, false)
		}
	}()

	return output, nil
}

func (inet Inet) listenSeparately(ctx context.Context) (<-chan infra.Conn, error) {
	output := make(chan infra.Conn)

	go func() {
		defer close(output)

		for ifaceName := range ip.Interfaces(ctx) {
			go func(ifaceName string) {
				for conn := range listenInterface(ctx, ifaceName) {
					output <- conn
				}
			}(ifaceName)
		}
	}()

	return output, nil
}

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
		hostPort := net.JoinHostPort(ip, strconv.Itoa(defaultListenPort))

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
