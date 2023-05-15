package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"io"
)

// launch launches an app using apphost
func launch(w io.ReadWriter, node _node.Node, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: launch <runtime> <app_path>")
	}

	host, err := modules.Find[*apphost.Module](node.Modules())
	if err != nil {
		return errors.New("apphost module not found")
	}

	return host.Launch(args[0], args[1])
}
