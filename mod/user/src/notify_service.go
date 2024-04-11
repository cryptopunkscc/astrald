package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"strings"
	"time"
)

const notifyServiceName = "shares.notify"

type NotifyService struct {
	*Module
}

func (srv *NotifyService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var userID = query.Target()

	if query.Caller().IsEqual(srv.node.Identity()) {
		return net.Accept(query, caller, func(conn net.SecureConn) {
			conn.Close()

			relays, err := srv.relay.FindExternalRelays(userID)
			if err != nil {
				return
			}

			for _, relay := range relays {
				relay := relay
				go func() {
					var q = net.NewQuery(userID, userID, notifyServiceName+"?"+query.Caller().PublicKeyHex())
					var hints = net.DefaultHints().WithValue(router.ViaRouterHintKey, srv.relay.RouterFuncVia(relay))

					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					conn, err := net.RouteWithHints(ctx, srv.node.Router(), q, hints)
					if err == nil {
						conn.Close()
					}
				}()
			}
		})
	}

	if query.Caller().IsEqual(query.Target()) {
		if i := strings.IndexByte(query.Query(), '?'); i != -1 {
			realCaller, err := id.ParsePublicKeyHex(query.Query()[i+1:])
			if err == nil {
				remoteShare, err := srv.shares.FindRemoteShare(userID, realCaller)
				if err != nil {
					return net.Reject()
				}

				return net.Accept(query, caller, func(conn net.SecureConn) {
					conn.Close()
					remoteShare.Sync(context.TODO())
				})
			}
		}
	}

	remoteShare, err := srv.shares.FindRemoteShare(userID, query.Caller())
	if err != nil {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		conn.Close()
		remoteShare.Sync(context.TODO())
	})
}
