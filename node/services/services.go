package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"time"
)

type Services interface {
	Register(ctx context.Context, identity id.Identity, name string, handler query.Router) (*Service, error)
	RouteQuery(ctx context.Context, query query.Query, w net.SecureWriteCloser) (net.SecureWriteCloser, error)
	Find(name string) (*Service, error)
	List() []ServiceInfo
}

type ServiceInfo struct {
	Name         string
	Identity     id.Identity
	RegisteredAt time.Time
}
