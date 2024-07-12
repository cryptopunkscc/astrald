package admin

import (
	"cmp"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"
)

var _ admin.Command = &CmdNode{}

type CmdNode struct {
	mod *Module
}

func (cmd *CmdNode) Exec(term admin.Terminal, args []string) error {
	var nodeID = cmd.mod.node.Identity()

	term.Printf("%v (%v)\n\n", nodeID, admin.Faded(nodeID.PublicKeyHex()))

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
	term.Printf("%s: %s\n", admin.Header("Modules"), strings.Join(names, " "))
	if coreNode, ok := cmd.mod.node.(*core.CoreNode); ok {
		term.Printf("%s: %v\n", admin.Header("Uptime"), time.Since(coreNode.StartedAt()).Round(time.Second))
	}

	term.Printf("\n%s\n\n", admin.Header("Endpoints"))

	for _, endpoint := range cmd.mod.node.Infra().Endpoints() {
		term.Printf("%-8s %v\n", endpoint.Network(), endpoint)
	}

	// Show routes
	term.Printf("\n%s\n\n", admin.Header("Routes"))
	var routeFmt = "%-32s %-32s\n"
	var routes = cmd.mod.node.LocalRouter().Routes()
	slices.SortFunc(routes, func(a, b node.LocalRoute) int {
		return cmp.Compare(a.Name, b.Name)
	})

	term.Printf(routeFmt, admin.Header("Name"), admin.Header("Type"))
	for _, route := range routes {
		t := reflect.TypeOf(route.Target)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		term.Printf(routeFmt, route.Name, admin.Keyword(t.String()))
	}

	return nil
}

func (cmd *CmdNode) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: node\n")
	return nil
}

func (cmd *CmdNode) ShortDescription() string {
	return "show node info"
}
