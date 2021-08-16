package etor

// Embedded TOR network support

import (
	"context"
	"github.com/cretz/bine/tor"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
)

const networkName = "tor"

type driver struct {
	tor *tor.Tor
}

func NewDriver() *driver {
	_tor, err := tor.Start(nil, &tor.StartConf{DebugWriter: nil})
	if err != nil {
		panic(err)
	}

	return &driver{tor: _tor}
}

var _ net.UnicastNetwork = &driver{}

func (drv driver) Network() string {
	return networkName
}

func (drv *driver) Dial(ctx context.Context, addr net.Addr) (net.Conn, error) {
	dialer, err := drv.tor.Dialer(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	conn, err := dialer.Dial("tcp", addr.String())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return net.WrapConn(conn, true), nil
}

func (drv *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	onion, _ := drv.tor.Listen(nil, &tor.ListenConf{Version3: true, RemotePorts: []int{1791}})

	output := make(chan net.Conn)

	log.Println("started listening on", onion.ID+".onion:1791")

	// Accept connections from the network
	go func() {
		for {
			conn, err := onion.Accept()
			if err != nil {
				onion.Close()
				log.Println("stopped listening on", onion.ID+".onion")
				break
			}
			output <- net.WrapConn(conn, false)
		}
	}()

	go func() {
		<-ctx.Done()
		onion.Close()
	}()

	return output, nil
}
