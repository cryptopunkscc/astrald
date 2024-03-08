package storage

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
)

type Handler interface {
	fmt.Stringer
	Handle(conn io.ReadWriter, args string) error
}

func RegisterHandler(ctx context.Context, handler Handler) error {
	return Register(ctx, handler.String(), handler.Handle)
}

type Handle func(conn io.ReadWriter, args string) error

func Register(ctx context.Context, name string, handle Handle) (err error) {
	listener, err := astral.Register(name)
	if err != nil {
		return
	}
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()
	go func() {
		for q := range listener.QueryCh() {
			go func(q *astral.QueryData) {
				var err error
				defer func() {
					if err != nil {
						log.Println(err)
					}
				}()
				conn, err := q.Accept()
				if err != nil {
					return
				}
				args := conn.Query()[len(name):]
				if err = handle(conn, args); err != nil {
					return
				}
			}(q)
		}
	}()

	return
}
