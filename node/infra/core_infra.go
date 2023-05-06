package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/config"
	"os"
	"strings"
	"sync"
)

var _ Infra = &CoreInfra{}

type CoreInfra struct {
	node           Node
	config         Config
	configStore    config.Store
	networkDrivers map[string]Driver
}

func NewCoreInfra(node Node, configStore config.Store) (*CoreInfra, error) {
	var i = &CoreInfra{
		node:           node,
		configStore:    configStore,
		networkDrivers: make(map[string]Driver),
		config:         defaultConfig,
	}

	if err := configStore.LoadYAML(configName, &i.config); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Error("config error: %s", err)
		} else {
			log.Errorv(2, "config error: %s", err)
		}
	}

	// load network drivers
	if err := i.loadDrivers(); err != nil {
		panic(err)
	}

	return i, nil
}

func (i *CoreInfra) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	log.Log("enabled drivers: %s", strings.Join(i.config.Drivers, " "))

	for _, name := range i.config.Drivers {
		if network, found := i.networkDrivers[name]; found {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := network.Run(ctx); err != nil {
					log.Error("network %s error: %s", name, err)
				} else {
					log.Log("network %s done", name)
				}
			}()
		} else {
			log.Error("network driver not found: %s", name)
		}
	}

	wg.Wait()

	return nil
}

func (i *CoreInfra) Node() Node {
	return i.node
}
