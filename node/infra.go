package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"io"
)

type Infra struct {
	// networks
	inet    *inet.Inet
	tor     *tor.Tor
	gateway *gw.Gateway
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
	case gw.NetworkName:
		return gw.Unpack(data)
	}
	return NewUnsupportedAddr(network, data), nil
}

func (infra *Infra) configure(node *Node) error {
	var err error

	cfg := &node.Config.Infra

	// Configure internet
	infra.inet = inet.New(cfg.Inet)

	err = astral.AddNetwork(infra.inet)
	if err != nil {
		return err
	}

	// Configure tor
	if cfg.Tor.DataDir == "" {
		cfg.Tor.DataDir = node.dataDir
	}

	infra.tor = tor.New(cfg.Tor)
	err = astral.AddNetwork(infra.tor)
	if err != nil {
		return err
	}

	// Configure astral gateways for mesh links
	infra.gateway = gw.New(&filteredQuerier{
		Querier:    node,
		FilteredID: node.Identity(),
	}, cfg.Gw)
	err = astral.AddNetwork(infra.gateway)
	if err != nil {
		return err
	}

	return nil
}

type filteredQuerier struct {
	gw.Querier
	FilteredID id.Identity
}

func (f filteredQuerier) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	if remoteID.IsEqual(f.FilteredID) {
		return nil, errors.New("filtered")
	}
	return f.Querier.Query(ctx, remoteID, query)
}
