package core

import (
	"context"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"sync"
)

var _ node.Infra = &CoreInfra{}

type CoreInfra struct {
	assets assets.Assets
	log    *log.Logger

	endpoints []node.EndpointLister
	dialers   map[string]node.Dialer
	unpackers map[string]node.Unpacker
	parsers   map[string]node.Parser
	mu        sync.RWMutex
}

func NewCoreInfra(assets assets.Assets, log *log.Logger) (*CoreInfra, error) {
	var i = &CoreInfra{
		assets:    assets,
		dialers:   make(map[string]node.Dialer),
		unpackers: make(map[string]node.Unpacker),
		parsers:   make(map[string]node.Parser),
		log:       log.Tag("infra"),
	}

	return i, nil
}

func (infra *CoreInfra) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}
