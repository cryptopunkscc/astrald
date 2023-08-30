package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"io"
	"sync"
)

type ConcurrentHandshake struct {
	localID  id.Identity
	remoteID id.Identity
	workers  int
}

func NewConcurrentHandshake(localID id.Identity, remoteID id.Identity, workers int) *ConcurrentHandshake {
	return &ConcurrentHandshake{localID: localID, remoteID: remoteID, workers: workers}
}

func (h *ConcurrentHandshake) Outbound(ctx context.Context, conns <-chan net.Conn) <-chan net.Link {
	var ch = make(chan net.Link, h.workers)
	var wg sync.WaitGroup

	// start handshake workers
	wg.Add(h.workers)
	for i := 0; i < h.workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return

				case conn, ok := <-conns:
					if !ok {
						return
					}

					hctx, _ := context.WithTimeout(ctx, HandshakeTimeout)
					l, err := link.Open(hctx, conn, h.remoteID, h.localID)

					// if handshake failed, try next connection
					if err != nil {
						switch {
						case errors.Is(err, io.EOF):
						case errors.Is(err, context.DeadlineExceeded):
						case errors.Is(err, context.Canceled):
						default:
							//TODO: add injectable error handler?
						}
						conn.Close()
						continue
					}

					ch <- l
					return
				}

			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
		for conn := range conns {
			conn.Close()
		}
	}()

	return ch
}
