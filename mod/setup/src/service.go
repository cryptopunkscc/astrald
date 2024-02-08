package setup

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/mod/setup"
	"github.com/cryptopunkscc/astrald/net"
)

type SetupService struct {
	*Module
}

func (srv *SetupService) Run(ctx context.Context) error {
	srv.node.LocalRouter().AddRoute(setup.ModuleName, srv)
	defer srv.node.LocalRouter().RemoveRoute(setup.ModuleName)

	srv.setDefaultIdentity()

	<-ctx.Done()

	return nil
}

func (srv *SetupService) Serve(conn net.SecureConn) {
	defer conn.Close()

	var d = NewSetupDialogue(srv.Module, conn)
	var err = d.start()
	if err != nil {
		d.Say("Error: %v", err)
	}

	srv.setDefaultIdentity()
}

func (srv *SetupService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if hints.Origin != net.OriginLocal {
		return net.Reject()
	}

	if len(srv.user.Identities()) > 0 {
		return net.Reject()
	}

	return net.Accept(query, caller, srv.Serve)
}

func (srv *SetupService) setDefaultIdentity() {
	if len(srv.user.Identities()) == 0 {
		srv.apphost.SetDefaultIdentity(srv.node.Identity())
		srv.presence.SetFlags(presence.SetupFlag)
		srv.presence.SetVisible(true)
	} else {
		srv.apphost.SetDefaultIdentity(id.Identity{})
		srv.presence.ClearFlags(presence.SetupFlag)
	}
}
