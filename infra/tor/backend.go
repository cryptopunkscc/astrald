package tor

import (
	"context"
	"net"
)

type Key []byte

type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() string
	PrivateKey() Key
}

type Backend interface {
	Run(context.Context, Config) error
	Dial(ctx context.Context, network string, addr string) (net.Conn, error)
	Listen(context.Context, Key) (Listener, error)
}

var backends = make(map[string]Backend)

func AddBackend(name string, backend Backend) {
	backends[name] = backend
}
