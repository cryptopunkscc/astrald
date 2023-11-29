package admin

import (
	. "github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/node/router"
	"reflect"
)

func (cmd *CmdNet) routes(term Terminal, _ []string) error {
	var routes []router.Route

	routes = cmd.mod.node.Router().Routes()

	var hf = "%-32s %-32s %-32s %6s\n"
	var rf = "%-32s %-32s %-32s %6d\n"

	term.Printf(hf, Header("Caller"), Header("Target"), Header("Type"), Header("Prio"))
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
