package presence

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"io"
)

const recentPath = "presence.recent"

type Provider struct {
	*Module
	*routers.PathRouter
}

func NewProvider(mod *Module) *Provider {
	p := &Provider{
		Module:     mod,
		PathRouter: routers.NewPathRouter(mod.node.Identity(), false),
	}

	p.AddRouteFunc(recentPath, p.recent)

	return p
}

func (mod *Provider) recent(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !mod.Auth.Authorize(q.Caller, presence.ActionList, nil) {
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

		list := mod.List()

		switch args.Format {
		case "json":
			err := json.NewEncoder(conn).Encode(list)
			if err != nil {
				mod.log.Errorv(2, "recent(): json encode: %v", err)
				return
			}

		default:
			for _, i := range list {
				_, err := i.WriteTo(conn)
				if err != nil {
					mod.log.Errorv(2, "recent(): write: %v", err)
					return
				}
			}
		}
	})
}
