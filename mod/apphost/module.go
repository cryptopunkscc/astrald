package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"net"
	"sync"
)

type Module struct {
	config    Config
	node      node.Node
	conns     <-chan net.Conn
	listeners []net.Listener
	tokens    map[string]string
}

func (mod *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var workerCount = mod.config.Workers

	mod.conns = mod.listen(ctx)

	log.Infov(2, "running %d workers", workerCount)

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wg.Done()
			if err := mod.worker(ctx); err != nil {
				log.Error("[%d] error: %s", i, err)
			}
		}(i)
	}

	wg.Wait()

	return nil
}
