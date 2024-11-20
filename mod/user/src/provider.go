package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"io"
)

type Provider struct {
	mod    *Module
	router *routers.PathRouter
}

func NewProvider(mod *Module) *Provider {
	p := &Provider{
		mod:    mod,
		router: routers.NewPathRouter(nil, false),
	}

	p.router.AddRouteFunc("user.nodes", p.Nodes)

	return p
}

func (p *Provider) Nodes(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	var args struct {
		Format string `query:"optional"`
		Names  bool   `query:"optional"`
	}
	_, err := query.ParseTo(q.Query, &args)
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		nodes := p.mod.Nodes(p.mod.userID)

		switch args.Format {
		case "json":
			if args.Names {
				var names []string
				for _, n := range nodes {
					names = append(names, p.mod.Dir.DisplayName(n))
				}
				err = json.NewEncoder(conn).Encode(names)
			} else {
				err = json.NewEncoder(conn).Encode(nodes)
			}

		default:
			for _, node := range nodes {
				_, err = node.WriteTo(conn)
				if err != nil {
					return
				}
			}
		}
	})
}
