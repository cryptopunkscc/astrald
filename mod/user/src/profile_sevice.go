package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"io"
)

type ProfileService struct {
	*Module
}

type ProfileData struct {
	Alias string `json:"alias"`
}

func (srv *ProfileService) RouteQuery(ctx context.Context, query astral.Query, caller io.WriteCloser, hints astral.Hints) (io.WriteCloser, error) {
	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		enc := json.NewEncoder(conn)
		enc.Encode(srv.getProfile(query.Target()))
	})
}

func (srv *ProfileService) getProfile(identity id.Identity) (p ProfileData) {
	p.Alias = srv.Dir.DisplayName(identity)

	return
}
