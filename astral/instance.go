package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

var instance Astral

func AddNetwork(network infra.Network) error {
	return instance.AddNetwork(network)
}

func LinkAt(localID id.Identity, remoteID id.Identity, addr infra.Addr) (*link.Link, error) {
	return instance.LinkAt(localID, remoteID, addr)
}

func Link(localID id.Identity, remoteID id.Identity, conn infra.Conn) (*link.Link, error) {
	return instance.Link(localID, remoteID, conn)
}

func Listen(ctx context.Context, localID id.Identity) (<-chan *link.Link, <-chan error) {
	return instance.Listen(ctx, localID)
}

func Unpack(networkName string, addr []byte) (infra.Addr, error) {
	return instance.Unpack(networkName, addr)
}

func NetworkNames() []string {
	return instance.NetworkNames()
}

func Addresses() []infra.AddrDesc {
	return instance.Addresses()
}

func Network(name string) infra.Network {
	return instance.Network(name)
}

func Announce(ctx context.Context, id id.Identity) error {
	return instance.Announce(ctx, id)
}

func Discover(ctx context.Context) (<-chan infra.Presence, error) {
	return instance.Discover(ctx)
}

func Dial(addr infra.Addr) (infra.Conn, error) {
	return instance.Dial(addr)
}
