package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
	"strings"
)

type Handler interface {
	fmt.Stringer
	Handle(conn io.ReadWriter, args *cslq.Decoder) error
}

func RegisterHandler(ctx context.Context, handler Handler) error {
	return Register(ctx, handler.String(), handler.Handle)
}

type Handle func(conn io.ReadWriter, args *cslq.Decoder) error

func Register(ctx context.Context, name string, handle Handle) (err error) {
	listener, err := astral.Register(name + "*")
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
				defer func() {
					if err != nil {
						log.Println(err)
					}
				}()
				conn, err := q.Accept()
				if err != nil {
					return
				}
				r := strings.NewReader(conn.Query()[len(name):])
				dec := cslq.NewDecoder(base64.NewDecoder(BaseEncoding, r))
				if err := handle(conn, dec); err != nil {
					return
				}
			}(q)
		}
	}()

	return
}
