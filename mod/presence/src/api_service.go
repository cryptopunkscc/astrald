package presence

import (
	"cmp"
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/net"
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

func (srv *APIService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !srv.mod.node.Auth().Authorize(query.Caller(), presence.ScanAction) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.Conn) {
		defer conn.Close()

		list := srv.mod.List()

		slices.SortFunc(list, func(a, b *presence.Presence) int {
			return cmp.Compare(a.Alias, b.Alias)
		})

		enc := json.NewEncoder(conn)
		enc.Encode(list)
	})
}
