package admin

import (
	"cmp"
	. "github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/router"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"
)

var _ Command = &CmdNode{}

type CmdNode struct {
	mod *Module
}

func (cmd *CmdNode) Exec(term Terminal, args []string) error {
	var nodeID = cmd.mod.node.Identity()

	term.Printf("%v (%v)\n\n", nodeID, Faded(nodeID.PublicKeyHex()))

	// Show modules
	var names []string
	for _, m := range cmd.mod.node.Modules().Loaded() {
		var t = reflect.TypeOf(m)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		n := strings.SplitN(t.String(), ".", 2)
		names = append(names, n[0])
	}
	sort.Strings(names)
	term.Printf("%s: %s\n", Header("Modules"), strings.Join(names, " "))
	if coreNode, ok := cmd.mod.node.(*node.CoreNode); ok {
		term.Printf("%s: %v\n", Header("Uptime"), time.Since(coreNode.StartedAt()).Round(time.Second))
	}

	// Show routes
	term.Printf("\n%s\n\n", Header("Routes"))
	var routeFmt = "%-32s %-32s\n"
	var routes = cmd.mod.node.Routes()
	slices.SortFunc(routes, func(a, b router.QueryRoute) int {
		return cmp.Compare(a.Name, b.Name)
	})

	term.Printf(routeFmt, Header("Name"), Header("Type"))
	for _, route := range routes {
		term.Printf(routeFmt, route.Name, reflect.TypeOf(route.Router))
	}

	return nil
}

func (cmd *CmdNode) help(term Terminal, _ []string) error {
	term.Printf("usage: node\n")
	return nil
}

func (cmd *CmdNode) ShortDescription() string {
	return "show node info"
}
