package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"sync"
)

const logTag = "infra"

var _ Infra = &CoreInfra{}

type CoreInfra struct {
	config Config
	assets assets.Assets
	log    *log.Logger

	endpoints []EndpointLister
	dialers   map[string]Dialer
	unpackers map[string]Unpacker
	parsers   map[string]Parser
	mu        sync.RWMutex
}

func NewCoreInfra(assets assets.Assets, log *log.Logger) (*CoreInfra, error) {
	var i = &CoreInfra{
		assets:    assets,
		dialers:   make(map[string]Dialer),
		unpackers: make(map[string]Unpacker),
		parsers:   make(map[string]Parser),
		config:    defaultConfig,
		log:       log.Tag(logTag),
	}

	// load config file
	_ = assets.LoadYAML(configName, &i.config)

	return i, nil
}

func (infra *CoreInfra) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}
