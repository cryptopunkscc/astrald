package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/lib/wrapper/embedded"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"sync"
	"time"
)

func serviceRunner(wg *sync.WaitGroup, name string, runners ...ServiceRunner) node.ModuleRunner {
	return embeddedServicesRunner{wg, name, runners}
}

type ServiceRunner interface {
	Run(ctx context.Context, api wrapper.Api) error
}

type embeddedServicesRunner struct {
	wg      *sync.WaitGroup
	name    string
	runners []ServiceRunner
}

func (e embeddedServicesRunner) String() string {
	return e.name
}

func (e embeddedServicesRunner) Run(ctx context.Context, n *node.Node) error {
	api := &embedded.Adapter{Ctx: ctx, Node: n}
	e.wg.Add(len(e.runners))
	for _, r := range e.runners {
		runner := r
		go func() {
			defer e.wg.Done()
			counter := 0
			var err error
			for counter < 10 {
				start := time.Now().UnixMilli()
				err = runner.Run(ctx, api)
				if err != nil || ctx.Err() != nil {
					log.Println("service", runner, "finished error:", err)
					return
				}
				end := time.Now().UnixMilli()
				if end-start < 2000 {
					counter++
				} else {
					counter = 0
				}
			}
			log.Println(runner, "aborted because failing constantly")
		}()
	}
	return nil
}
