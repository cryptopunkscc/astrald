package lan

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
	goNet "net"
	"strconv"
)

func (drv *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	output := make(chan net.Conn)

	// Get all our IP addresses
	addrs, err := goNet.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// skip non-private addresses
		if !(isAddrLocalNetwork(addr)) {
			continue
		}

		ip, _ := net.SplitIPMask(addr.String())
		hostPort := goNet.JoinHostPort(ip, strconv.Itoa(int(drv.port)))

		// Listen on the port
		l, err := goNet.Listen("tcp", hostPort)
		if err != nil {
			continue
		}

		log.Println("listening on", hostPort)

		// Accept connections from the network
		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					l.Close()
					log.Println("stopped listening on", hostPort)
					break
				}
				output <- net.WrapConn(conn, false)
			}
		}()

		go func() {
			<-ctx.Done()
			l.Close()
		}()
	}

	return output, nil
}
