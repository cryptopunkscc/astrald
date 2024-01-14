package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	_net "net"
)

var _ infra.Dialer = &Module{}

// Dial tries to establish a Driver connection to the provided address
func (mod *Module) Dial(ctx context.Context, endpoint net.Endpoint) (conn net.Conn, err error) {
	endpoint, err = mod.Unpack(endpoint.Network(), endpoint.Pack())
	if err != nil {
		return nil, err
	}

	var e = endpoint.(Endpoint)

	ctx, cancel := context.WithTimeout(ctx, mod.config.DialTimeout)
	defer cancel()

	var connCh = make(chan _net.Conn, 1)
	var errCh = make(chan error, 1)

	// Attempt a connection in the background
	go func() {
		defer close(connCh)
		defer close(errCh)

		c, err := mod.proxy.DialContext(ctx, "tcp", e.String())
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
		return newConn(c, e, true), nil
	case err = <-errCh:
		return nil, err
	case <-ctx.Done():
		err = ctx.Err()
		return nil, err
	}
}
