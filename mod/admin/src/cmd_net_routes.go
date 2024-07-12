package admin

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node"
	"reflect"
)

func (cmd *CmdNet) routes(term admin.Terminal, _ []string) error {
	var routes []node.Route

	routes = cmd.mod.node.Router().Routes()

	var hf = "%-32s %-32s %-32s %6s\n"
	var rf = "%-32s %-32s %-32s %6d\n"

	term.Printf(hf, admin.Header("Caller"), admin.Header("Target"), admin.Header("Type"), admin.Header("Prio"))
	for _, route := range routes {
		var t = reflect.TypeOf(route.Router)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		term.Printf(rf,
			route.Caller,
			route.Target,
			t.String(),
			route.Priority,
		)
	}

	return nil
}
