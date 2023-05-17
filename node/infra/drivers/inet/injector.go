package inet

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, assets assets.Store) error {
	drv := &Driver{
		config:      defaultConfig,
		infra:       i,
		publicAddrs: make([]net.Endpoint, 0),
	}

	if assets != nil {
		if err := assets.LoadYAML(DriverName, &drv.config); err != nil {
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
