package presence

import (
	"cmp"
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
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

func (srv *APIService) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	if !srv.mod.Auth.Authorize(query.Caller, presence.ScanAction, nil) {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		list := srv.mod.List()

		slices.SortFunc(list, func(a, b *presence.Presence) int {
			return cmp.Compare(a.Alias, b.Alias)
		})

		enc := json.NewEncoder(conn)
		enc.Encode(list)
	})
}
