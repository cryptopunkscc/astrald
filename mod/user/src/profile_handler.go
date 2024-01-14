package user

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

const userProfileServiceType = "user.profile"
const userProfileServiceName = "profile"

type ProfileHandler struct {
	*Module
}

type ProfileData struct {
	Alias string `json:"alias"`
}

func (handler *ProfileHandler) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		enc := json.NewEncoder(conn)
		enc.Encode(handler.getProfile(query.Target()))
	})
}

func (handler *ProfileHandler) getProfile(identity id.Identity) (p ProfileData) {
	p.Alias = handler.node.Resolver().DisplayName(identity)

	return
}
