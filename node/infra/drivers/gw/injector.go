package gw

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, assets assets.Store, l *log.Logger) error {
	drv := &Driver{
		infra:  i,
		config: defaultConfig,
		log:    l,
	}

	l.Root().PushFormatFunc(func(v any) ([]log.Op, bool) {
		ep, ok := v.(Endpoint)
		if !ok {
			return nil, false
		}

		var ops = make([]log.Op, 0)

		if format, ok := l.Render(ep.gate); ok {
			ops = append(ops, format...)
		} else {
			ops = append(ops, log.OpText{Text: ep.gate.String()})
		}

		ops = append(ops,
			log.OpColor{Color: log.White},
			log.OpText{Text: ":"},
			log.OpReset{},
		)

		if format, ok := l.Render(ep.target); ok {
			ops = append(ops, format...)
		} else {
			ops = append(ops, log.OpText{Text: ep.gate.String()})
		}

		return ops, true
	})

	_ = assets.LoadYAML(DriverName, &drv.config)

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
