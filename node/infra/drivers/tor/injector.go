package tor

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, configStore config.Store) error {
	drv := &Driver{
		config:      defaultConfig,
		configStore: configStore,
	}

	if configStore != nil {
		if err := configStore.LoadYAML(DriverName, &drv.config); err != nil {
			log.Errorv(2, "error reading config: %s", err)
		}
	}

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
