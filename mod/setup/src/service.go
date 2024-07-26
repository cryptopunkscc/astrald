package setup

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/setup"
)

type SetupService struct {
	*Module
}

func (srv *SetupService) Run(ctx context.Context) error {
	srv.AddRoute(setup.ModuleName, srv)
	defer srv.RemoveRoute(setup.ModuleName)

	srv.setDefaultIdentity()

	if srv.needsSetup() {
		srv.Presence.SetVisible(true)
	}

	<-ctx.Done()

	return nil
}

func (srv *SetupService) Serve(conn astral.Conn) {
	defer conn.Close()

	var d = NewSetupDialogue(srv.Module, conn)
	var err = d.start()
	if err != nil {
		d.Say("Error: %v", err)
	}

	srv.setDefaultIdentity()
	srv.Presence.Broadcast() // update our setup flag
}

func (srv *SetupService) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	if hints.Origin != astral.OriginLocal {
		return astral.Reject()
	}

	if !srv.needsSetup() {
		return astral.Reject()
	}

	return astral.Accept(query, caller, srv.Serve)
}

func (srv *SetupService) setDefaultIdentity() {
	if srv.User.UserID().IsZero() {
		srv.Apphost.SetDefaultIdentity(srv.node.Identity())
	} else {
		srv.Apphost.SetDefaultIdentity(id.Identity{})
	}
}
