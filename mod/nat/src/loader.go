package nat

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, l *log.Logger) (core.Module, error) {
	mod := &Module{
		node: node,
		log:  l,
		cond: sync.NewCond(&sync.Mutex{}),
	}

	mod.pool = NewPairPool(mod)
	mod.ops.AddStructPrefix(mod, "Op")

	return mod, nil
}

func init() {
	err := core.RegisterModule(nat.ModuleName, Loader{})
	if err != nil {
		panic(err)
	}
}
