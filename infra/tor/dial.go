package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

// Dial tries to establish a Tor connection to the provided address
func (tor Tor) Dial(ctx context.Context, addr infra.Addr) (conn infra.Conn, err error) {
	ctx, cancel := context.WithTimeout(ctx, tor.config.getDialTimeout())
	defer cancel()

	// Convert to Tor address
	torAddr, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	var connCh = make(chan net.Conn, 1)
	var errCh = make(chan error, 1)

	// Attempt a connection in the background
	go func() {
		defer close(connCh)
		defer close(errCh)

		c, err := tor.proxy.Dial("tcp", addr.String())
		if err != nil {
			errCh <- err
			return
		}

		// Return the connection if we're still waiting for it, close it otherwise
		select {
		case connCh <- c:
		default:
			c.Close()
		}
	}()

	// Wait for the first result
	select {
	case c := <-connCh:
		return newConn(c, torAddr, true), nil
	case err = <-errCh:
		return nil, err
	case <-ctx.Done():
		err = ctx.Err()
		return nil, err
	}
}
