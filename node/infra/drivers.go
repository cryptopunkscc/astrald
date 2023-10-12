package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
)

var drivers map[string]DriverInjector

type Driver interface {
	// Run the network driver
	Run(ctx context.Context) error
}

type DriverInjector interface {
	Inject(Infra, assets.Store, *log.Logger) error
}

func RegisterDriver(name string, driver DriverInjector) error {
	if drivers == nil {
		drivers = make(map[string]DriverInjector)
	}
	if _, found := drivers[name]; found {
		return errors.New("driver already added")
	}

	drivers[name] = driver

	return nil
}

func (infra *CoreInfra) AddDriver(name string, driver Driver) error {
	if _, found := infra.networkDrivers[name]; found {
		return errors.New("driver already added")
	}

	infra.networkDrivers[name] = driver
	return nil
}

func (infra *CoreInfra) Drivers() map[string]Driver {
	enabledDrivers := make(map[string]Driver)
	for k, v := range infra.networkDrivers {
		if infra.config.driversContain(k) {
			enabledDrivers[k] = v
		}
	}
	return enabledDrivers
}

func (infra *CoreInfra) loadDrivers() error {
	for name, injector := range drivers {
		if err := injector.Inject(infra, infra.assets, infra.log.Tag(name)); err != nil {
			infra.log.Errorv(1, "error loading network driver %s: %s", name, err)
		}
	}

	return nil
}

func GetDriver[T any](infra Infra, name string) (T, bool) {
	if driver, found := infra.Drivers()[name]; found {
		d, ok := driver.(T)
		return d, ok
	} else {
		var none T
		return none, false
	}
}
