package apphost

import (
	"net"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ apphost.Module = &Module{}

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	db     *DB
	router routing.OpRouter

	listeners []net.Listener
	conns     <-chan net.Conn
	handlers  sig.Set[*QueryHandler]
	enRoute   sig.Map[astral.Nonce, *queryEnRoute]
}

func (mod *Module) Run(ctx *astral.Context) error {
	var wg sync.WaitGroup
	var workerCount = mod.config.Workers

	mod.conns = mod.listen(ctx)

	// spawn workers
	mod.log.Logv(2, "spawning %v workers", workerCount)
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer debug.SaveLog(debug.SigInt)

			defer wg.Done()
			if err := mod.worker(ctx); err != nil {
				mod.log.Error("[worker:%v] error: %v", i, err)
			}
		}(i)
	}

	// start the object server
	objectServer := NewHTTPServer(mod)
	objectServer.Run(ctx)

	wg.Wait()

	return nil
}

func (mod *Module) Router() astral.Router {
	return &mod.router
}

func (mod *Module) String() string {
	return apphost.ModuleName
}

func (mod *Module) RoutingPriority() int {
	return astral.RoutingPriorityHigh
}
