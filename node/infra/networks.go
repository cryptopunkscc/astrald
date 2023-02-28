package infra

import (
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

var networkPriorities map[string]int

// NetworkPriority returns network's priority
func NetworkPriority(netName string) int {
	return networkPriorities[netName]
}

func (i *Infra) setupNetworks() error {
	var err error

	// inet
	if i.config.networksContain(inet.NetworkName) {
		i.inet, err = inet.New(i.config.Inet, i.localID)
		if err == nil {
			i.addNetwork(inet.NetworkName, i.inet)
		} else {
			log.Error("inet: %s", err)
		}
	}

	// tor
	if i.config.networksContain(tor.NetworkName) {
		i.tor, err = tor.New(i.config.Tor, i.rootDir)
		if err == nil {
			i.addNetwork(tor.NetworkName, i.tor)
		} else {
			log.Error("tor: %s", err)
		}
	}

	// gateway
	if i.config.networksContain(gw.NetworkName) {
		i.gateway, err = gw.New(i.config.Gw, i.Querier)
		if err == nil {
			i.addNetwork(gw.NetworkName, i.gateway)
		} else {
			log.Error("gw: %s", err)
		}
	}

	// bluetooth
	if i.config.networksContain(bt.NetworkName) {
		if bt.Instance != nil {
			i.bluetooth = bt.Instance
			i.addNetwork(bt.NetworkName, i.bluetooth)
		} else {
			log.Error("bt: adapter unavailable")
		}
	}

	return nil
}

func (i *Infra) addNetwork(name string, n infra.Network) {
	i.networks[name] = n
}

func init() {
	networkPriorities = map[string]int{
		inet.NetworkName: 400,
		bt.NetworkName:   300,
		gw.NetworkName:   200,
		tor.NetworkName:  100,
	}
}
