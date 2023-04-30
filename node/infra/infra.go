package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/link"
	"os"
	"strings"
	"sync"
)

type Querier interface {
	Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error)
}

type Infra struct {
	Querier  Querier
	networks map[string]infra.Network
	localID  id.Identity
	rootDir  string

	gateways  []infra.AddrSpec
	inet      *inet.Inet
	tor       *tor.Tor
	gateway   *gw.Gateway
	bluetooth bt.Client
	config    *Config
	logLevel  int
}

var log = _log.Tag("infra")

func New(localID id.Identity, configStore config.Store, querier Querier, rootDir string) (*Infra, error) {
	var i = &Infra{
		localID:  localID,
		Querier:  querier,
		rootDir:  rootDir,
		gateways: make([]infra.AddrSpec, 0),
		networks: make(map[string]infra.Network),
		config:   &Config{},
	}

	if err := configStore.LoadYAML(configName, i.config); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error("config error: %s", err)
		} else {
			log.Errorv(2, "config error: %s", err)
		}
	}

	// static gateways
	if err := i.setupStaticGateways(); err != nil {
		log.Error("setupStaticGateways: %s", err)
	}

	// setup networks
	if err := i.setupNetworks(); err != nil {
		log.Error("setupNetworks: %s", err)
	}

	return i, nil
}

func (i *Infra) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	list := make([]string, 0, len(i.networks))
	for n := range i.networks {
		list = append(list, n)
	}
	log.Log("enabled networks: %s", strings.Join(list, " "))

	for name, net := range i.networks {
		name, net := name, net
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := net.Run(ctx); err != nil {
				log.Error("network %s error: %s", name, err)
			} else {
				log.Log("network %s done", name)
			}
		}()
	}

	wg.Wait()

	return nil
}

func (i *Infra) Networks() map[string]infra.Network {
	clone := make(map[string]infra.Network)
	for k, v := range i.networks {
		clone[k] = v
	}
	return clone
}

func (i *Infra) LocalAddrs() []infra.AddrSpec {
	list := make([]infra.AddrSpec, 0)

	type addrLister interface {
		Addresses() []infra.AddrSpec
	}

	// collect addresses from all networks that support it
	for _, net := range i.networks {
		if lister, ok := net.(addrLister); ok {
			list = append(list, lister.Addresses()...)
		}
	}

	// append gateways
	list = append(list, i.gateways...)

	return list
}

func (i *Infra) Gateway() *gw.Gateway {
	return i.gateway
}

func (i *Infra) Bluetooth() bt.Client {
	return i.bluetooth
}

func (i *Infra) Tor() *tor.Tor {
	return i.tor
}

func (i *Infra) Inet() *inet.Inet {
	return i.inet
}

func (i *Infra) Config() *Config {
	return i.config
}

func (i *Infra) setupStaticGateways() error {
	for _, gate := range i.config.Gateways {
		gateID, err := id.ParsePublicKeyHex(gate)
		if err != nil {
			log.Error("error parsing gateway %s: %s", gate, err.Error())
			continue
		}

		i.gateways = append(i.gateways, infra.AddrSpec{
			Addr:   gw.NewAddr(gateID, i.localID),
			Global: true,
		})
	}
	return nil
}
