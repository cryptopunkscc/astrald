package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"os"
	"strings"
	"sync"
)

const logTag = "infra"

var _ Infra = &CoreInfra{}

type CoreInfra struct {
	node           Node
	config         Config
	assets         assets.Store
	networkDrivers map[string]Driver
	log            *log.Logger
}

func NewCoreInfra(node Node, assets assets.Store, log *log.Logger) (*CoreInfra, error) {
	var i = &CoreInfra{
		node:           node,
		assets:         assets,
		networkDrivers: make(map[string]Driver),
		config:         defaultConfig,
		log:            log.Tag(logTag),
	}

	if err := assets.LoadYAML(configName, &i.config); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			i.log.Error("config error: %s", err)
		} else {
			i.log.Errorv(2, "config error: %s", err)
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

	i.log.Log("enabled drivers: %s", strings.Join(i.config.Drivers, " "))

	for _, name := range i.config.Drivers {
		if network, found := i.networkDrivers[name]; found {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := network.Run(ctx); err != nil {
					i.log.Error("network %s error: %s", name, err)
				} else {
					i.log.Log("network %s done", name)
				}
			}()
		} else {
			i.log.Error("network driver not found: %s", name)
		}
	}

	wg.Wait()

	return nil
}

func (i *CoreInfra) Node() Node {
	return i.node
}
