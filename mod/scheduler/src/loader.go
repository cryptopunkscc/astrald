package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/sig"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	mod := &Module{
		node: node,
		log:  l,
		q:    &sig.Queue[scheduler.Action]{},
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(scheduler.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
