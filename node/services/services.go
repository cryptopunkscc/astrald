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
	Find(identity id.Identity, name string) (*Service, error)
	FindByName(name string) []*Service
	List() []ServiceInfo
}

type ServiceInfo struct {
	Name         string
	Identity     id.Identity
	RegisteredAt time.Time
}
