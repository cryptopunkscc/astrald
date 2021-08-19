package apphost

import (
	"context"
	"fmt"
	"log"
	"net"
)

const tcpPort = 8625

func serveTCP(ctx context.Context) (<-chan net.Conn, error) {
	outCh := make(chan net.Conn)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", tcpPort))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(outCh)

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		addr := listener.Addr().String()

		log.Println("apps: listen tcp", addr)

		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			outCh <- conn
		}

		log.Println("apps: closed tcp", addr)
	}()

	return outCh, nil
}
