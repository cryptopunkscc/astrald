package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"strconv"
	"strings"
)

var portConfig = net.ListenConfig{}

func (inet *Inet) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	ctx, cancel := context.WithCancel(ctx)

	// start the listener
	var addrStr = ":" + strconv.Itoa(inet.ListenPort())
	listener, err := portConfig.Listen(ctx, "tcp", addrStr)
	if err != nil {
		return nil, err
	}

	var output = make(chan infra.Conn)
	go func() {
		<-ctx.Done()
		log.Log("stop listen tcp %s", addrStr)
		listener.Close()
		close(output)
	}()

	log.Log("listen tcp %s", log.Em(addrStr))

	go func() {
		defer cancel()
		for {
			conn, err := listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Error("accept: %s", err)
				}
				return
			}

			output <- newConn(conn, false)
		}
	}()

	return output, nil
}
