package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"log"
	"net"
	"strconv"
	"strings"
)

func (inet Inet) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	ctx, cancel := context.WithCancel(ctx)
	// start the listener
	var addrStr = ":" + strconv.Itoa(inet.listenPort)
	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		return nil, err
	}

	var output = make(chan infra.Conn)
	go func() {
		<-ctx.Done()
		close(output)
		listener.Close()
		log.Println("[inet] stop listen tcp", addrStr)
	}()

	log.Println("[inet] listen tcp", addrStr)

	go func() {
		defer cancel()
		for {
			conn, err := listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("[inet] accept error:", err)
				}
				return
			}

			output <- newConn(conn, false)
		}
	}()

	return output, nil
}
