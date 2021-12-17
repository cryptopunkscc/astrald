package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/storage"
	"io"
	"log"
)

type Querier interface {
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

var _ infra.Unpacker = &Infra{}

type Infra struct {
	Querier  Querier
	Store    storage.Store
	networks map[string]infra.Network

	gateways []infra.AddrSpec
	inet     *inet.Inet
	tor      *tor.Tor
	gateway  *gw.Gateway
}

func New(cfg config.Infra, querier Querier, store storage.Store) (*Infra, error) {
	i := &Infra{
		Querier:  querier,
		Store:    store,
		gateways: make([]infra.AddrSpec, 0),
	}
	var err error

	for _, gate := range cfg.Gateways {
		addr, err := gw.Parse(gate)
		if err != nil {
			log.Println("error parsing gateway:", err)
			continue
		}
		i.gateways = append(i.gateways, infra.AddrSpec{
			Addr:   addr,
			Global: true,
		})
	}

	// Configure internet
	i.inet = inet.New(cfg.Inet)

	err = i.addNetwork(i.inet)
	if err != nil {
		return nil, err
	}

	i.tor = tor.New(cfg.Tor, i.Store)
	err = i.addNetwork(i.tor)
	if err != nil {
		return nil, err
	}

	// Configure astral gateways for mesh links
	i.gateway = gw.New(cfg.Gw, i.Querier)
	err = i.addNetwork(i.gateway)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (i *Infra) Networks() <-chan infra.Network {
	ch := make(chan infra.Network, len(i.networks))
	for _, n := range i.networks {
		ch <- n
	}
	close(ch)
	return ch
}

func (i *Infra) Gateways() []infra.AddrSpec {
	return i.gateways
}

func (i *Infra) Addresses() []infra.AddrSpec {
	list := make([]infra.AddrSpec, 0)

	list = append(list, i.inet.Addresses()...)
	list = append(list, i.tor.Addresses()...)
	list = append(list, i.Gateways()...)

	return list
}

func (i *Infra) addNetwork(n infra.Network) error {
	if i.networks == nil {
		i.networks = make(map[string]infra.Network)
	}
	i.networks[n.Name()] = n
	return nil
}
