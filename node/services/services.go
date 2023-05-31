package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"time"
)

type Services interface {
	Register(ctx context.Context, identity id.Identity, name string, handler HandlerFunc) (*Service, error)
	Query(ctx context.Context, caller id.Identity, query string, link *link.Link) (*Conn, error)
	Find(name string) (*Service, error)
	List() []ServiceInfo
}

type ServiceInfo struct {
	Name         string
	Identity     id.Identity
	RegisteredAt time.Time
}
