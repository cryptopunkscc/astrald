package gateway

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

const forwardServiceFormat = "gateway.forward.%s"

type ForwardService struct {
	*Module
	target id.Identity
	router net.Router
}

func (srv *ForwardService) Run(ctx context.Context) error {
	serviceName := fmt.Sprintf(forwardServiceFormat, srv.target.PublicKeyHex())

	_, err := srv.node.Services().Register(ctx, srv.node.Identity(), serviceName, srv)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (srv *ForwardService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var fwdQuery = net.NewQueryNonce(srv.node.Identity(), srv.target, query.Query(), query.Nonce())

	return srv.router.RouteQuery(ctx, fwdQuery, caller, hints.SetAllowRedirect())
}
