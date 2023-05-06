package inet

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, configStore config.Store) error {
	drv := &Driver{
		config:      defaultConfig,
		infra:       i,
		publicAddrs: make([]net.Endpoint, 0),
	}

	if configStore != nil {
		if err := configStore.LoadYAML(DriverName, &drv.config); err != nil {
			log.Errorv(2, "error reading config: %s", err)
		}
	}

	// Add public addresses
	for _, addrStr := range drv.config.PublicAddr {
		addr, err := Parse(addrStr)
		if err != nil {
			log.Error("parse error: %s", err)
			continue
		}

		drv.publicAddrs = append(drv.publicAddrs, addr)
		log.Log("public addr: %s", addr)
	}

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
