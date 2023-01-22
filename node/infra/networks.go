package infra

import (
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
)

func (i *Infra) setupNetworks(cfg config.Infra) error {
	var err error

	// inet
	if i.config.IsNetworkEnabled(inet.NetworkName) {
		i.inet, err = inet.New(cfg.Inet, i.localID)
		if err == nil {
			i.addNetwork(inet.NetworkName, i.inet)
		} else {
			i.Logf(log.Normal, "inet error: %s", err)
		}
	}

	// tor
	if i.config.IsNetworkEnabled(tor.NetworkName) {
		i.tor, err = tor.New(cfg.Tor, i.rootDir)
		if err == nil {
			i.addNetwork(tor.NetworkName, i.tor)
		} else {
			i.Logf(log.Normal, "tor error: %s", err)
		}
	}

	// gateway
	if i.config.IsNetworkEnabled(gw.NetworkName) {
		i.gateway, err = gw.New(cfg.Gw, i.Querier)
		if err == nil {
			i.addNetwork(gw.NetworkName, i.gateway)
		} else {
			i.Logf(log.Normal, "gw error: %s", err)
		}
	}

	// bluetooth
	if i.config.IsNetworkEnabled(bt.NetworkName) {
		if bt.Instance != nil {
			i.bluetooth = bt.Instance
			i.addNetwork(bt.NetworkName, i.bluetooth)
		} else {
			i.Logf(log.Normal, "bluetooth error: adapter unavailable", err)
		}
	}

	return nil
}

func (i *Infra) addNetwork(name string, n infra.Network) {
	i.networks[name] = n
}
