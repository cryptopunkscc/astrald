package status

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/status"
	"io"
)

const listPath = "status.list"

type Provider struct {
	*Module
	*routers.PathRouter
}

func NewProvider(mod *Module) *Provider {
	p := &Provider{
		Module:     mod,
		PathRouter: routers.NewPathRouter(mod.node.Identity(), false),
	}

	p.AddRouteFunc(listPath, p.list)

	return p
}

func (mod *Provider) list(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !mod.Auth.Authorize(q.Caller, status.ActionList, nil) {
		return query.Reject()
	}

	var args struct {
		Format string `query:"optional"`
	}

	_, err := query.ParseTo(q.Query, &args)
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		list := mod.Broadcasters()

		switch args.Format {
		case "json":
			err := json.NewEncoder(conn).Encode(list)
			if err != nil {
				mod.log.Errorv(2, "list(): json encode: %v", err)
				return
			}

		default:
			for _, i := range list {
				_, err := i.WriteTo(conn)
				if err != nil {
					mod.log.Errorv(2, "list(): write: %v", err)
					return
				}
			}
		}
	})
}
