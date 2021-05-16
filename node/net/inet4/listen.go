package inet4

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/net"
	goNet "net"
)

func Listen(ctx context.Context, port int) (<-chan net.Conn, error) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)

	// Listen on the port
	tcpListen, err := goNet.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}

	accept := make(chan net.Conn)
	output := make(chan net.Conn)

	// Accept connections from the network
	go func() {
		defer close(accept)
		for {
			conn, err := tcpListen.Accept()
			if err != nil {
				break
			}
			accept <- Wrap(conn, false)
		}
	}()

	// Pass new connections over and shut down when context is done
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				_ = tcpListen.Close()
				return
			case c := <-accept:
				output <- c
			}
		}
	}()

	return output, nil
}
