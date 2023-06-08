package tor

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, assets assets.Store, l *log.Logger) error {
	drv := &Driver{
		config: defaultConfig,
		assets: assets,
		log:    l,
	}

	if assets != nil {
		assets.LoadYAML(DriverName, &drv.config)
	}

	l.Root().PushFormatFunc(func(v any) ([]log.Op, bool) {
		ep, ok := v.(Endpoint)
		if !ok {
			return nil, false
		}

		return []log.Op{
			log.OpColor{Color: log.Cyan},
			log.OpText{Text: ep.String()},
			log.OpReset{},
		}, true
	})

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
