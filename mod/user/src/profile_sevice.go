package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
)

const userProfileServiceType = "user.profile"
const userProfileServiceName = "profile"

type ProfileService struct {
	*Module
}

type ProfileData struct {
	Alias string `json:"alias"`
}

func (srv *ProfileService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.Conn) {
		defer conn.Close()

		enc := json.NewEncoder(conn)
		enc.Encode(srv.getProfile(query.Target()))
	})
}

func (srv *ProfileService) getProfile(identity id.Identity) (p ProfileData) {
	p.Alias = srv.node.Resolver().DisplayName(identity)

	return
}
