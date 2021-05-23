package net

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/auth/id"
)

type Driver interface {
	Network() string
	Dial(ctx context.Context, addr Addr) (Conn, error)
	Listen(ctx context.Context) (<-chan Conn, error)
	Advertise(ctx context.Context) error
	Scan(ctx context.Context) (<-chan *Ad, error)
}

type Ad struct {
	Identity id.Identity
	Addr     Addr
}
