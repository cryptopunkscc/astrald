package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
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
	endpoints      []EndpointLister
	dialers        map[string]Dialer
	unpackers      map[string]Unpacker
	mu             sync.Mutex
}

func NewCoreInfra(node Node, assets assets.Store, log *log.Logger) (*CoreInfra, error) {
	var i = &CoreInfra{
		node:           node,
		assets:         assets,
		networkDrivers: make(map[string]Driver),
		dialers:        make(map[string]Dialer),
		unpackers:      make(map[string]Unpacker),
		config:         defaultConfig,
		log:            log.Tag(logTag),
	}

	// load config file
	_ = assets.LoadYAML(configName, &i.config)

	// load network drivers
	if err := i.loadDrivers(); err != nil {
		panic(err)
	}

	return i, nil
}

func (infra *CoreInfra) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	var loaded []string
	for name := range infra.networkDrivers {
		loaded = append(loaded, name)
	}

	infra.log.Log("drivers loaded: %s, enabled: %s",
		strings.Join(loaded, " "),
		strings.Join(infra.config.Drivers, " "),
	)

	for _, name := range infra.config.Drivers {
		name := name
		if network, found := infra.networkDrivers[name]; found {
			wg.Add(1)
			go func() {
				defer debug.SaveLog(func(p any) {
					infra.log.Error("network driver %s panicked: %v", name, p)
					debug.SigInt(p)
				})

				defer wg.Done()

				if err := network.Run(ctx); err != nil {
					infra.log.Error("network %s error: %s", name, err)
				} else {
					infra.log.Logv(1, "network %s done", name)
				}
			}()
		} else {
			infra.log.Error("network driver not found: %s", name)
		}
	}

	wg.Wait()

	return nil
}

func (infra *CoreInfra) Node() Node {
	return infra.node
}
