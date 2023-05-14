package bt

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, configStore config.Store) error {
	drv := &Driver{
		config: defaultConfig,
	}

	if configStore != nil {
		configStore.LoadYAML(DriverName, &drv.config)
	}

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
