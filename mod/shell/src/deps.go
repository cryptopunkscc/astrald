package shell

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Objects objects.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	if cnode, ok := mod.node.(*core.Node); ok {
		var added []any
		for _, m := range cnode.Modules().Loaded() {
			if m == mod {
				continue
			}

			if s, ok := m.(ops.HasOps); ok {
				mod.root.AddSet(getName(s), s.GetOpSet())
				added = append(added, m)
			}
		}
		if len(added) > 0 {
			mod.log.Logv(2, "shell scopes: %v"+strings.Repeat(", %v", len(added)-1), added...)
		}
	}

	return
}

func getName(v any) string {
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}

	if reflect.TypeOf(v).Kind() == reflect.Pointer {
		return reflect.TypeOf(v).Elem().String()
	}

	return reflect.TypeOf(v).String()
}
