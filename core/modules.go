package core

import (
	"github.com/cryptopunkscc/astrald/node"
)

func Load[M any](node node.Node, name string) (M, error) {
	mod, ok := node.Modules().Find(name).(M)
	if !ok {
		return mod, ModuleUnavailable(name)
	}
	return mod, nil
}
