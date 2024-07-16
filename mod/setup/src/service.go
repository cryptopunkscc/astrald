package setup

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
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

	if srv.needsSetup() {
		srv.presence.SetVisible(true)
	}

	<-ctx.Done()

	return nil
}

func (srv *SetupService) Serve(conn net.Conn) {
	defer conn.Close()

	var d = NewSetupDialogue(srv.Module, conn)
	var err = d.start()
	if err != nil {
		d.Say("Error: %v", err)
	}

	srv.setDefaultIdentity()
	srv.presence.Broadcast() // update our setup flag
}

func (srv *SetupService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if hints.Origin != net.OriginLocal {
		return net.Reject()
	}

	if !srv.needsSetup() {
		return net.Reject()
	}

	return net.Accept(query, caller, srv.Serve)
}

func (srv *SetupService) setDefaultIdentity() {
	if srv.user.UserID().IsZero() {
		srv.apphost.SetDefaultIdentity(srv.node.Identity())
	} else {
		srv.apphost.SetDefaultIdentity(id.Identity{})
	}
}
