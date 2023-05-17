package tor

import (
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, assets assets.Store) error {
	drv := &Driver{
		config: defaultConfig,
		assets: assets,
	}

	if assets != nil {
		assets.LoadYAML(DriverName, &drv.config)
	}

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
