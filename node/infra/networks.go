package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"strings"
)

func (i *Infra) setupNetworks(ctx context.Context, cfg config.Infra) error {
	// inet
	if err := i.setupNetwork(func() (infra.Network, error) {
		return inet.New(cfg.Inet, i.localID)
	}); err != nil {
		i.Logf(log.Normal, "inet error: %s", err)
	}

	// tor
	if err := i.setupNetwork(func() (infra.Network, error) {
		return tor.Run(ctx, cfg.Tor, i.Store)
	}); err != nil {
		i.Logf(log.Normal, "tor error: %s", err)
	}

	// gateway
	if err := i.setupNetwork(func() (infra.Network, error) {
		return gw.New(cfg.Gw, i.Querier)
	}); err != nil {
		i.Logf(log.Normal, "gw error: %s", err)
	}

	// bluetooth
	if err := i.setupNetwork(func() (infra.Network, error) {
		return bt.Instance, nil
	}); err != nil {
		i.Logf(log.Normal, "bt error: %s", err)
	}

	list := make([]string, 0, len(i.networks))
	for n := range i.networks {
		list = append(list, n)
	}

	i.Logf(log.Normal, "configured networks: %s", strings.Join(list, ", "))

	return nil
}

func (i *Infra) setupNetwork(f func() (infra.Network, error)) error {
	network, err := f()
	if err != nil {
		return err
	}
	return i.addNetwork(network)
}

func (i *Infra) addNetwork(n infra.Network) error {
	if n == nil {
		return errors.New("network is nil")
	}
	if i.networks == nil {
		i.networks = make(map[string]infra.Network)
	}
	i.networks[n.Name()] = n
	return nil
}
