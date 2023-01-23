package nat

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/node/link"
	"io"
	"time"
)

const cmdInit = "init"
const cmdPing = "ping"
const cmdAddr = "addr"
const cmdGo = "go"
const cmdTime = "time"

const maxTimeDistance = 30 * time.Second

func (mod *Module) runServer(ctx context.Context) error {
	port, err := mod.node.Ports.Register(portName)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case query, ok := <-port.Queries():
			if !ok {
				return nil
			}

			if mod.mapping.extAddr.IsZero() {
				query.Reject()
				continue
			}

			if query.IsLocal() {
				query.Reject()
				continue
			}

			conn, err := query.Accept()
			if err != nil {
				continue
			}

			go func() {
				if err := mod.serve(ctx, conn); err != nil {
					switch {
					case errors.Is(err, io.EOF):
					default:
						log.Error("serve: %s", err)
					}
				}
			}()

		}
	}
}

func (mod *Module) serve(ctx context.Context, conn *hub.Conn) error {
	defer conn.Close()
	var c = cslq.NewEndec(conn)
	var remoteAddr inet.Addr
	var cmd string

	for {
		var err error
		if err = c.Decode("[c]c", &cmd); err != nil {
			return err
		}

		switch cmd {
		case cmdInit:
			if err = c.Encode("x00"); err != nil {
				return err
			}

		case cmdPing:
			if err = c.Encode("x00"); err != nil {
				return err
			}

		case cmdAddr:
			var buf []byte
			if err = c.Decode("[c]c", &buf); err != nil {
				return err
			}
			remoteAddr, err = inet.Unpack(buf)
			if err != nil {
				return err
			}

			if err = c.Encode("[c]c", mod.mapping.extAddr.Pack()); err != nil {
				return err
			}

		case cmdGo:
			if remoteAddr.IsZero() {
				return errors.New("protocol error: go before addr")
			}
			authed, err := mod.makeLink(ctx, remoteAddr, id.Identity{})
			if err != nil {
				continue
			}

			return mod.node.Peers.AddLink(link.NewFromConn(authed))

		case cmdTime:
			if remoteAddr.IsZero() {
				return errors.New("protocol error: go before addr")
			}

			var startTime int

			if err = c.Decode("q", &startTime); err != nil {
				return err
			}

			waitTime := time.Until(time.Unix(int64(startTime), 0))
			if waitTime <= 0 {
				continue
			}
			if waitTime > maxTimeDistance {
				continue
			}

			<-time.After(waitTime)

			authed, err := mod.makeLink(ctx, remoteAddr, id.Identity{})
			if err != nil {
				continue
			}

			return mod.node.Peers.AddLink(link.NewFromConn(authed))

		default:
			return errors.New("protocol error: unknown request type")
		}
	}
}
