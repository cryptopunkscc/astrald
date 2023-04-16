package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/inet"
)

type Infra interface {
	Dial(ctx context.Context, addr infra.Addr) (conn infra.Conn, err error)
	LocalAddrs() []infra.AddrSpec
	Unpack(network string, data []byte) (infra.Addr, error)
	Inet() *inet.Inet
}
