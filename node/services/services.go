package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"time"
)

type Services interface {
	Register(ctx context.Context, name string) (*Service, error)
	RegisterAs(ctx context.Context, name string, identity id.Identity) (*Service, error)
	Query(ctx context.Context, query string, link *link.Link) (*Conn, error)
	QueryAs(ctx context.Context, query string, link *link.Link, identity id.Identity) (*Conn, error)
	Release(name string) error
	List() []ServiceInfo
}

type ServiceInfo struct {
	Name         string
	Identity     id.Identity
	RegisteredAt time.Time
}
