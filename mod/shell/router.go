package shell

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

var _ astral.Router = &Router{}

type Router struct {
	scope *Scope
	log   *log.Logger
}

func NewRouter(scope *Scope, log *log.Logger) *Router {
	return &Router{scope: scope, log: log}
}

func (r Router) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	path, params := query.Parse(q.Query)

	if !r.scope.Exists(path) {
		return query.RouteNotFound(r)
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		ctx := astral.WrapContext(ctx, q.Caller)
		env := NewBinaryEnv(conn, conn)

		var err = r.scope.Call(ctx, env, path, params)
		if err != nil {
			if r.log != nil {
				r.log.Errorv(1, "error calling %v for %v: %v", path, q.Caller, err)
			}
		}
	})
}
