package node

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/infra"
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

type Infra struct {
	// networks
	inet   *inet.Inet
	tor    *tor.Tor
	astral *iastral.Astral
}

func (infra *Infra) Addresses() []infra.AddrSpec {
	return astral.Addresses()
}

func (infra *Infra) UnpackAddr(network string, data []byte) (infra.Addr, error) {
	switch network {
	case tor.NetworkName:
		return tor.Unpack(data)
	case inet.NetworkName:
		return inet.Unpack(data)
	case iastral.NetworkName:
		return iastral.Unpack(data)
	}
	return NewUnsupportedAddr(network, data), nil
}

func (infra *Infra) configure(node *Node) error {
	var err error

	// Configure internet
	infra.inet = inet.New(node.Config.Infra.Inet)

	err = astral.AddNetwork(infra.inet)
	if err != nil {
		return err
	}

	// Configure tor
	if node.Config.Infra.Tor.DataDir == "" {
		node.Config.Infra.Tor.DataDir = node.dataDir
	}

	infra.tor = tor.New(node.Config.Infra.Tor)
	err = astral.AddNetwork(infra.tor)
	if err != nil {
		return err
	}

	// Configure astral for mesh links
	infra.astral = iastral.NewAstral(node, node.Config.Infra.Astral)
	err = astral.AddNetwork(infra.astral)
	if err != nil {
		return err
	}

	return nil
}
