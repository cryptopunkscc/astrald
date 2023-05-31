package reflectlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"sync"
)

const serviceName = "net.reflectlink"

type Module struct {
	node node.Node
	log  *log.Logger
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
