package inet

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/infra"
	"strconv"
)

var _ infra.DriverInjector = &Injector{}

type Injector struct{}

func (*Injector) Inject(i infra.Infra, assets assets.Store, l *log.Logger) error {
	drv := &Driver{
		config:      defaultConfig,
		infra:       i,
		log:         l,
		publicAddrs: make([]net.Endpoint, 0),
	}

	if assets != nil {
		if err := assets.LoadYAML(DriverName, &drv.config); err != nil {
			l.Errorv(2, "error reading config: %s", err)
		}
	}

	// Add public addresses
	for _, addrStr := range drv.config.PublicAddr {
		addr, err := Parse(addrStr)
		if err != nil {
			l.Error("error parsing '%s': %s", addrStr, err)
			continue
		}

		drv.publicAddrs = append(drv.publicAddrs, addr)
	}

	l.Root().PushFormatFunc(func(v any) ([]log.Op, bool) {
		ep, ok := v.(Endpoint)
		if !ok {
			return nil, false
		}

		var ops = make([]log.Op, 0)

		ip := ep.ip.String()
		if ep.ver == ipv6 {
			ip = "[" + ip + "]"
		}

		ops = append(ops,
			log.OpColor{Color: log.Cyan},
			log.OpText{Text: ip},
			log.OpReset{},
		)

		if ep.port != 0 {
			ops = append(ops,
				log.OpColor{Color: log.White},
				log.OpText{Text: ":"},
				log.OpReset{},
				log.OpColor{Color: log.Cyan},
				log.OpText{Text: strconv.Itoa(int(ep.port))},
				log.OpReset{},
			)
		}

		return ops, true
	})

	return i.AddDriver(DriverName, drv)
}

func init() {
	if err := infra.RegisterDriver(DriverName, &Injector{}); err != nil {
		panic(err)
	}
}
