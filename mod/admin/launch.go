package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
)

// launch launches an app using apphost
func launch(w io.ReadWriter, node _node.Node, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: launch <runtime> <app_path>")
	}

	coreNode, ok := node.(*_node.CoreNode)
	if !ok {
		return errors.New("not running on a core node")
	}

	mod := coreNode.Modules().FindModule("apphost")
	if mod == nil {
		return errors.New("apphost module not found")
	}

	host, ok := mod.(*apphost.Module)
	if !ok {
		return errors.New("apphost module not found")
	}

	return host.Launch(args[0], args[1])
}
