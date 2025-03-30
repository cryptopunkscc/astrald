package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
)

func (mod *Module) worker(ctx context.Context) error {
	for conn := range mod.conns {
		var done = make(chan struct{})

		go func() {
			select {
			case <-ctx.Done():
				conn.Close()
			case <-done:
			}
		}()

		err := NewSession(mod, conn, mod.log).Serve(astral.NewContext(ctx).WithIdentity(mod.node.Identity()))

		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
		case strings.Contains(err.Error(), "connection closed"):
		case strings.Contains(err.Error(), "use of closed network connection"):
		case strings.Contains(err.Error(), "read/write on closed pipe"):
		case strings.Contains(err.Error(), "session ended"):
		default:
			mod.log.Error("serve error: %v", err)
		}

		conn.Close()
		close(done)
	}
	return nil
}
