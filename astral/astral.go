package astral

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

type Astral struct {
	networks map[string]infra.Network
}

func (astral *Astral) AddNetwork(network infra.Network) error {
	if len(network.Name()) == 0 {
		return errors.New("invalid network name")
	}
	if astral.networks == nil {
		astral.networks = make(map[string]infra.Network)
	}
	if _, found := astral.networks[network.Name()]; found {
		return errors.New("network already added")
	}
	astral.networks[network.Name()] = network
	return nil
}

func (astral *Astral) Networks() <-chan infra.Network {
	ch := make(chan infra.Network, len(astral.networks))
	for _, network := range astral.networks {
		ch <- network
	}
	close(ch)
	return ch
}

func (astral *Astral) NetworkNames() []string {
	names := make([]string, 0, len(astral.networks))
	for name := range astral.networks {
		names = append(names, name)
	}
	return names
}

func (astral *Astral) Addresses() []infra.AddrDesc {
	list := make([]infra.AddrDesc, 0)

	for _, network := range astral.networks {
		list = append(list, network.Addresses()...)
	}

	return list
}

func (astral *Astral) Network(name string) infra.Network {
	return astral.networks[name]
}

func (astral *Astral) Link(localID id.Identity, remoteID id.Identity, addr infra.Addr) (*link.Link, error) {
	// sanity check
	if localID.IsEqual(remoteID) {
		return nil, errors.New("cannot link with self")
	}

	conn, err := astral.dial(addr)
	if err != nil {
		return nil, err
	}

	authConn, err := auth.HandshakeOutbound(context.Background(), conn, remoteID, localID)
	if err != nil {
		return nil, err
	}

	link := link.New(authConn)

	return link, nil
}

func (astral *Astral) Unpack(networkName string, addr []byte) (infra.Addr, error) {
	if astral.networks == nil {
		return nil, infra.ErrUnsupportedNetwork
	}

	network, found := astral.networks[networkName]
	if !found {
		return nil, infra.ErrUnsupportedNetwork
	}

	return network.Unpack(addr)
}

func (astral *Astral) dial(addr infra.Addr) (infra.Conn, error) {
	network, found := astral.networks[addr.Network()]
	if !found {
		return nil, infra.ErrUnsupportedNetwork
	}
	return network.Dial(context.Background(), addr)
}
