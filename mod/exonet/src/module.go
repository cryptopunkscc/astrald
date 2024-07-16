package exonet

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

var _ exonet.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	assets resources.Resources

	dialers   sig.Map[string, exonet.Dialer]
	unpackers sig.Map[string, exonet.Unpacker]
	parser    sig.Map[string, exonet.Parser]
	resolvers sig.Set[exonet.Resolver]
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (mod *Module) Resolve(ctx context.Context, identity id.Identity) ([]exonet.Endpoint, error) {
	var res sig.Set[exonet.Endpoint]

	var wg sync.WaitGroup
	for _, r := range mod.resolvers.Clone() {
		r := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := r.Resolve(ctx, identity)
			if err != nil {
				return
			}
			res.Add(v...)
		}()
	}
	wg.Wait()

	return res.Clone(), nil
}

func (mod *Module) AddResolver(resolver exonet.Resolver) {
	if resolver != nil {
		mod.resolvers.Add(resolver)
	}
}

func (mod *Module) Dial(ctx context.Context, endpoint exonet.Endpoint) (conn exonet.Conn, err error) {
	d, found := mod.dialers.Get(endpoint.Network())
	if found {
		return d.Dial(ctx, endpoint)
	}

	return nil, exonet.ErrUnsupportedNetwork
}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	u, found := mod.unpackers.Get(network)
	if found {
		return u.Unpack(network, data)
	}

	return nil, exonet.ErrUnsupportedNetwork
}

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	p, found := mod.parser.Get(network)
	if found {
		return p.Parse(network, address)
	}

	return nil, exonet.ErrUnsupportedNetwork

}

func (mod *Module) SetDialer(network string, dialer exonet.Dialer) {
	mod.dialers.Replace(network, dialer)
}

func (mod *Module) SetUnpacker(network string, unpacker exonet.Unpacker) {
	mod.unpackers.Replace(network, unpacker)
}

func (mod *Module) SetParser(network string, parser exonet.Parser) {
	mod.parser.Replace(network, parser)
}
