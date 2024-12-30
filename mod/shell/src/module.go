package shell

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
	"io"
)

var _ shell.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	*shell.Router
	root shell.Scope
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Root() *shell.Scope {
	return &mod.root
}

func (mod *Module) serve(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		var s = NewSession(mod, conn)

		err := s.Run(astral.WrapContext(ctx, q.Caller))
		if err != nil {
			mod.log.Errorv(2, "session with %v ended in error: %v", q.Caller, err)
		}
	})
}

type rw struct {
	astral.ObjectReader
	astral.ObjectWriter
}

func (mod *Module) String() string {
	return shell.ModuleName
}
