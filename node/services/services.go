package services

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type Services interface {
	net.Router
	Register(ctx context.Context, identity id.Identity, name string, handler net.Router) (*Service, error)
	Find(name string) (*Service, error)
	List() []ServiceInfo
}

type ServiceInfo struct {
	Name         string
	Identity     id.Identity
	RegisteredAt time.Time
}
