package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/config"
)

var drivers map[string]DriverInjector

type Driver interface {
	// Run the network driver
	Run(ctx context.Context) error
}

type DriverInjector interface {
	Inject(infra Infra, configStore config.Store) error
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

func (i *CoreInfra) AddDriver(name string, driver Driver) error {
	if _, found := i.networkDrivers[name]; found {
		return errors.New("driver already added")
	}

	i.networkDrivers[name] = driver
	return nil
}

func (i *CoreInfra) Drivers() map[string]Driver {
	enabledDrivers := make(map[string]Driver)
	for k, v := range i.networkDrivers {
		if i.config.driversContain(k) {
			enabledDrivers[k] = v
		}
	}
	return enabledDrivers
}

func (i *CoreInfra) loadDrivers() error {
	for name, injector := range drivers {
		if err := injector.Inject(i, i.configStore); err != nil {
			log.Errorv(1, "error loading network driver %s: %s", name, err)
		} else {
			log.Infov(2, "loaded network driver %s", name)
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
