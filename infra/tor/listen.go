package tor

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/tor/tc"
	"log"
	"net"
	"time"
)

const retryInterval = 15 * time.Second
const pingInterval = time.Minute

func (tor *Tor) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	output := make(chan infra.Conn)

	// Set up the listener for incoming tor connections
	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	// close the listener when context is done
	go func() {
		<-ctx.Done()
		tcpListener.Close()
	}()

	// process incoming connections
	go func() {
		defer close(output)
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				return
			}
			output <- newConn(conn, Addr{}, false)
		}
	}()

	// keep the onion up
	go func() {
		key, err := tor.getPrivateKey()
		if err != nil {
			return
		}

		for {
			err := tor.runOnion(ctx, key, tor.config.getListenPort(), tcpListener.Addr().String())
			if err != nil {
				log.Println("(tor) listen error:", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(retryInterval):
			}
		}
	}()

	return output, nil
}

func (tor *Tor) runOnion(ctx context.Context, key string, port int, target string) error {
	ctl, err := tor.connect()
	if err != nil {
		return err
	}
	defer ctl.Close()

	onion, err := ctl.AddOnion(key, tc.Port(port, target))
	if err != nil {
		return err
	}

	defer ctl.DelOnion(onion.ServiceID)

	tor.serviceAddr, _ = Parse(onion.ServiceID)

	log.Printf("(tor) listen %s\n", tor.serviceAddr.String())

	for {
		select {
		case <-time.After(pingInterval):
			ctl.GetConf("ControlPort") // ping the daemon
		case <-ctx.Done():
			return ctx.Err()
		case <-ctl.WaitClose():
			return errors.New("connection lost")
		}
	}
}

func (tor *Tor) connect() (*tc.Control, error) {
	return tc.Open(tc.Config{ControlAddr: tor.config.ControlAddr})
}
