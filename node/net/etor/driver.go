package etor

// Embedded TOR network support

import (
	"context"
	"fmt"
	_tor "github.com/cretz/bine/tor"
	"github.com/cryptopunkscc/astrald/node/net"
	"log"
)

type driver struct {
	tor *_tor.Tor
}

var _ net.UnicastNetwork = &driver{}

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
	output := make(chan net.Conn)

	// Accept connections from the network
	go func() {
		if drv.tor == nil {
			log.Println("initializing embedded tor...")
			tor, err := _tor.Start(nil, &_tor.StartConf{DebugWriter: nil})
			if err != nil {
				fmt.Println("tor error:", err)
				return
			}
			drv.tor = tor
		}

		onion, err := drv.tor.Listen(nil, &_tor.ListenConf{Version3: true, RemotePorts: []int{1791}})
		if err != nil {
			fmt.Println("tor error:", err)
			return
		}

		torURL := onion.ID + ".onion:1791"
		log.Println("listen tor", torURL)

		for {
			conn, err := onion.Accept()
			if err != nil {
				log.Println("closed tor", torURL)
				break
			}
			output <- net.WrapConn(conn, false)
		}

		go func() {
			<-ctx.Done()
			onion.Close()
		}()
	}()

	return output, nil
}

func init() {
	if err := net.AddUnicastNetwork("tor", &driver{}); err != nil {
		panic(err)
	}
}
