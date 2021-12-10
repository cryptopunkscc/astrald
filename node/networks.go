package node

import (
	"github.com/cryptopunkscc/astrald/astral"
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

func (node *Node) addNetworks() error {
	var err error

	// Configure internet
	node.inet = inet.New(node.Config.Net.Inet)

	err = astral.AddNetwork(node.inet)
	if err != nil {
		return err
	}

	// Configure tor
	node.tor = tor.New(node.Config.Net.Tor)
	err = astral.AddNetwork(node.tor)
	if err != nil {
		return err
	}

	// Configure astral for mesh links
	node.astral = iastral.NewAstral(node, node.Config.Net.Astral)
	err = astral.AddNetwork(node.astral)
	if err != nil {
		return err
	}

	return nil
}
