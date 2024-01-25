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
	lastSync, err := srv.LastSynced(query.Target(), query.Caller())
	if err != nil {
		return net.Reject()
	}

	if lastSync.IsZero() {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		conn.Close()

		srv.Sync(query.Target(), query.Caller())
	})
}
