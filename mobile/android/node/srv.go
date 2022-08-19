package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/lib/wrapper/embedded"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"time"
)

func serviceRunner(name string, runners ...ServiceRunner) node.ModuleRunner {
	return embeddedServicesRunner{name, runners}
}

type ServiceRunner interface {
	Run(ctx context.Context, api wrapper.Api) error
}

type embeddedServicesRunner struct {
	name    string
	runners []ServiceRunner
}

func (e embeddedServicesRunner) String() string {
	return e.name
}

func (e embeddedServicesRunner) Run(ctx context.Context, n *node.Node) error {
	api := &embedded.Adapter{Ctx: ctx, Node: n}
	for _, r := range e.runners {
		runner := r
		go func() {
			counter := 0
			var err error
			for counter < 10 {
				start := time.Now().UnixMilli()
				err = runner.Run(ctx, api)
				end := time.Now().UnixMilli()
				if end-start < 2000 {
					counter++
				} else {
					counter = 0
				}
				log.Println(runner, "run error:", err)
			}
			log.Println(runner, "aborted because failing constantly")
		}()
	}
	return nil
}
