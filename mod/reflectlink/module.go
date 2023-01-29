package reflectlink

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"sync"
)

const portName = "net.reflectlink"

type Module struct {
	node *node.Node
}

var log = _log.Tag(ModuleName)

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
