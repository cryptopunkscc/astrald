package reflectlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
	"sync"
)

const serviceName = "net.reflectlink"

type Module struct {
	node node.Node
	log  *log.Logger
}

func (mod *Module) Discover(ctx context.Context, caller id.Identity, medium string) ([]proto.ServiceEntry, error) {
	if medium == services.SourceLocal {
		return []proto.ServiceEntry{{
			Name:  serviceName,
			Type:  "net.reflectlink",
			Extra: nil,
		}}, nil
	}

	return []proto.ServiceEntry{}, nil
}

func (mod *Module) Run(ctx context.Context) error {
	ctx, abort := context.WithCancel(ctx)
	defer abort()

	var wg = sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		mod.runServer(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		mod.runClient(ctx)
	}()

	wg.Wait()
	return nil
}
