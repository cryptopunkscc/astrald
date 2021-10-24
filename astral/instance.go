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

func Link(localID id.Identity, remoteID id.Identity, addr infra.Addr) (*link.Link, error) {
	return instance.Link(localID, remoteID, addr)
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
