package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	_net "net"
	"strconv"
	"strings"
)

var _ infra.Listener = &Driver{}

var portConfig = _net.ListenConfig{}

func (drv *Driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	ctx, cancel := context.WithCancel(ctx)

	// start the listener
	var addrStr = ":" + strconv.Itoa(drv.ListenPort())
	listener, err := portConfig.Listen(ctx, "tcp", addrStr)
	if err != nil {
		cancel()
		return nil, err
	}

	var output = make(chan net.Conn)
	go func() {
		<-ctx.Done()
		drv.log.Log("stop listen tcp %s", addrStr)
		listener.Close()
		close(output)
	}()

	drv.log.Log("listen tcp %s", addrStr)

	go func() {
		defer cancel()
		for {
			conn, err := listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					drv.log.Error("accept: %s", err)
				}
				return
			}

			output <- newConn(conn, false)
		}
	}()

	return output, nil
}

func (drv *Driver) ListenPort() int {
	return drv.config.ListenPort
}
