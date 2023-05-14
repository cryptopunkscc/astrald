package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Services interface {
	Register(name string) (*Service, error)
	RegisterContext(ctx context.Context, name string) (*Service, error)
	Query(ctx context.Context, query string, link *link.Link) (*Conn, error)
}
