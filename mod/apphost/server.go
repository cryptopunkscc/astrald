package apphost

import (
	"context"
	"log"
	"net"
	"sync"
)

// Name of the control socket
const unixSocketName = "ctl.sock"

func Serve(ctx context.Context) <-chan net.Conn {
	outCh := make(chan net.Conn)

	go func() {
		defer close(outCh)

		var wg sync.WaitGroup

		unixConns, err := serveUnix(ctx)
		if err != nil {
			log.Println("apps: listen unix error:", err)
		} else {
			wg.Add(1)
			go func() {
				for c := range unixConns {
					outCh <- c
				}
				wg.Done()
			}()
		}

		tcpConns, err := serveTCP(ctx)
		if err != nil {
			log.Println("apps: listen unix error:", err)
		} else {
			wg.Add(1)
			go func() {
				for c := range tcpConns {
					outCh <- c
				}
				wg.Done()
			}()
		}

		wg.Wait()
	}()

	return outCh
}
