package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"net"
	"os"
	"sync"
)

type Module struct {
	config    Config
	node      node.Node
	conns     <-chan net.Conn
	log       *log.Logger
	listeners []net.Listener
	tokens    map[string]AppInfo
}

func (mod *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var workerCount = mod.config.Workers

	mod.conns = mod.listen(ctx)

	mod.log.Infov(2, "running %d workers", workerCount)

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wg.Done()
			if err := mod.worker(ctx); err != nil {
				mod.log.Error("[%d] error: %s", i, err)
			}
		}(i)
	}

	if len(mod.config.Boot) > 0 {
		mod.log.Infov(1, "starting boot apps...")
	}

	_, err := modules.WaitReady[*contacts.Module](ctx, mod.node.Modules())
	if err != nil {
		mod.log.Errorv(1, "error waiting for mod_contacts: %s", err)
	}

	for _, boot := range mod.config.Boot {
		boot := boot
		go func() {
			mod.log.Infov(1, "starting %s...", boot.App)
			err := mod.Launch(boot.App, boot.Args, os.Environ())
			if err != nil {
				mod.log.Infov(1, "app %s ended with error: %s", boot.App, err)
			}
		}()
	}

	wg.Wait()

	return nil
}
