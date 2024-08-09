package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

type ProfileService struct {
	*Module
}

type ProfileData struct {
	Alias string `json:"alias"`
}

func (srv *ProfileService) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		enc := json.NewEncoder(conn)
		enc.Encode(srv.getProfile(q.Target))
	})
}

func (srv *ProfileService) getProfile(identity *astral.Identity) (p ProfileData) {
	p.Alias = srv.Dir.DisplayName(identity)

	return
}
