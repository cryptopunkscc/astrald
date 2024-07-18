package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/astral"
)

const userProfileServiceType = "user.profile"
const userProfileServiceName = "profile"

type ProfileService struct {
	*Module
}

type ProfileData struct {
	Alias string `json:"alias"`
}

func (srv *ProfileService) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		enc := json.NewEncoder(conn)
		enc.Encode(srv.getProfile(query.Target()))
	})
}

func (srv *ProfileService) getProfile(identity id.Identity) (p ProfileData) {
	p.Alias = srv.dir.DisplayName(identity)

	return
}
