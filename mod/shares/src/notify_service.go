package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

const notifyServiceName = "shares.notify"

type NotifyService struct {
	*Module
}

func NewNotifyService(module *Module) *NotifyService {
	return &NotifyService{Module: module}
}

func (srv *NotifyService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(notifyServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(notifyServiceName)
	<-ctx.Done()
	return nil
}

func (srv *NotifyService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	remoteShare, err := srv.FindRemoteShare(query.Target(), query.Caller())
	if err != nil {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		conn.Close()
		remoteShare.Sync()
	})
}
