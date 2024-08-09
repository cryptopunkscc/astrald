package presence

import (
	"cmp"
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"io"
	"slices"
)

const scanServiceName = "presence.scan"

type APIService struct {
	mod *Module
}

func NewAPIService(mod *Module) *APIService {
	return &APIService{mod: mod}
}

func (srv *APIService) Run(ctx context.Context) error {
	srv.mod.AddRoute(scanServiceName, srv)
	<-ctx.Done()
	return nil
}

func (srv *APIService) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !srv.mod.Auth.Authorize(q.Caller, presence.ScanAction, nil) {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		list := srv.mod.List()

		slices.SortFunc(list, func(a, b *presence.Presence) int {
			return cmp.Compare(a.Alias, b.Alias)
		})

		enc := json.NewEncoder(conn)
		enc.Encode(list)
	})
}
