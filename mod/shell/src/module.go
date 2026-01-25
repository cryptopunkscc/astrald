package shell

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
)

var _ shell.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	root   ops.Set
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !q.Target.IsEqual(mod.node.Identity()) {
		return query.RouteNotFound(mod)
	}

	return mod.root.RouteQuery(ctx, q, w)
}

func (mod *Module) Root() *ops.Set {
	return &mod.root
}

func (mod *Module) String() string {
	return shell.ModuleName
}

func (mod *Module) NewLogAction(message string) shell.LogAction {
	return LogAction{
		mod:     mod,
		message: message,
	}
}
