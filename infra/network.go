package infra

import "context"

type Network interface {
	Name() string
	Unpack([]byte) (Addr, error)
	Dial(ctx context.Context, addr Addr) (Conn, error)
	Listen(ctx context.Context) (<-chan Conn, <-chan error)
	Advertise(ctx context.Context, payload []byte) <-chan error
	Scan(ctx context.Context) (<-chan Ad, <-chan error)
	Addresses() []AddrDesc
}

type Ad struct {
	Source  Addr
	Payload []byte
}
