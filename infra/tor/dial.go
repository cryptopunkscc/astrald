package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"strconv"
	"time"
)

// Dial tries to establish a Tor connection to the provided address
func (tor Tor) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	// Convert to Tor address
	torAddr, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	var connCh = make(chan net.Conn)
	var errCh = make(chan error)

	// Attempt a connection in the background
	go func() {
		defer close(connCh)
		defer close(errCh)

		conn, err := tor.proxy.Dial("tcp", addr.String()+":"+strconv.Itoa(defaultListenPort))
		if err != nil {
			errCh <- err
			return
		}

		// Return the connection if we're still waiting for it, close it otherwise
		select {
		case connCh <- conn:
		default:
			conn.Close()
		}
	}()

	// Wait for the first result
	select {
	case conn := <-connCh:
		return newConn(conn, torAddr, true), nil
	case err := <-errCh:
		return nil, err
	case <-time.After(tor.config.getDialTimeout()):
		return nil, infra.ErrDialTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
