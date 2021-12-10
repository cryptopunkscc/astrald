package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

var instance Astral

func AddNetwork(network infra.Network) error {
	return instance.AddNetwork(network)
}

func Listen(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	return instance.Listen(ctx)
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

// Dial connects to the address
func Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	return instance.Dial(ctx, addr)
}

// DialMany dials every address from addrCh and returns all successful connections via the output channel.
// Concurrency specifies how many concurrent dialers should be running and has to be at least 1, otherwise no attempt
// will be made and the output channel will close instantly.
func DialMany(ctx context.Context, addrCh <-chan infra.Addr, concurrency int) <-chan infra.Conn {
	return instance.DialMany(ctx, addrCh, concurrency)
}
