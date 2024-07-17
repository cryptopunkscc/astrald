package admin

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"reflect"
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
	if cnode, ok := cmd.mod.node.(*core.Node); ok {
		var names []string
		for _, m := range cnode.Modules().Loaded() {
			var t = reflect.TypeOf(m)
			for t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			n := strings.SplitN(t.String(), ".", 2)
			names = append(names, n[0])
		}
		sort.Strings(names)
		term.Printf("%s: %s\n", admin.Header("Modules"), strings.Join(names, " "))
		term.Printf("%s: %v\n", admin.Header("Uptime"), time.Since(cnode.StartedAt()).Round(time.Second))
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
