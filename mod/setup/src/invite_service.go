package setup

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	id2 "github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type InviteService struct {
	*Module
	Accept id2.Filter
}

func NewInviteService(module *Module, accept id2.Filter) *InviteService {
	return &InviteService{Module: module, Accept: accept}
}

func (srv *InviteService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute("setup.invite", srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute("setup.invite")

	<-ctx.Done()
	return nil
}

func (srv *InviteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !srv.needsSetup() {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.Conn) {
		defer conn.Close()
		var err error
		var cert relay.Cert

		err = cslq.Decode(conn, "v", &cert)
		if err != nil {
			srv.log.Errorv(2, "invite: error reading certificate: %v", err)
			return
		}

		if cert.ExpiresAt.Before(time.Now()) {
			srv.log.Errorv(2, "invite: received an expired cert")
			return
		}

		if !cert.RelayID.IsEqual(srv.node.Identity()) {
			srv.log.Errorv(2, "invite: certificate's relayID mismatch: %v", cert.RelayID)
			return
		}

		err = cert.VerifyTarget()
		if err != nil {
			srv.log.Errorv(2, "invite: invalid target signature: %v", err)
			return
		}

		if !srv.Accept(cert.TargetID) {
			srv.log.Info("invitation from %v rejected", cert.TargetID)
			return
		}
		srv.log.Info("invitation from %v accepted", cert.TargetID)

		cert.RelaySig, err = srv.keys.Sign(srv.node.Identity(), cert.Hash())
		if err != nil {
			srv.log.Error("invite: error signing certificate: %v", err)
			return
		}

		err = cert.Validate()
		if err != nil {
			srv.log.Error("invite: signed certificate invalid: %v", err)
		}

		err = srv.joinByCert(&cert)
		if err != nil {
			srv.log.Error("invite: joinByCert: %v", err)
			return
		}

		cslq.Encode(conn, "v", cert)

		srv.presence.Broadcast() // broadcast our presence to remove the setup flag

		return
	})
}

func (srv *InviteService) Invite(ctx context.Context, userID id2.Identity, nodeID id2.Identity) error {
	var err error
	var cert = relay.Cert{
		TargetID:  userID,
		RelayID:   nodeID,
		Direction: relay.Both,
		ExpiresAt: time.Now().Add(relay.DefaultCertDuration),
	}

	cert.TargetSig, err = srv.keys.Sign(cert.TargetID, cert.Hash())
	if err != nil {
		return fmt.Errorf("error signing invite certificate: %w", err)
	}

	err = cert.VerifyTarget()
	if err != nil {
		return fmt.Errorf("invite: signature verification failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	var query = net.NewQuery(userID, nodeID, "setup.invite")
	conn, err := net.Route(ctx, srv.node.Router(), query)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = cslq.Encode(conn, "v", &cert)
	if err != nil {
		return err
	}

	err = cslq.Decode(conn, "v", &cert)
	if err != nil {
		return err
	}

	err = cert.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (srv *InviteService) joinByCert(cert *relay.Cert) error {
	err := cert.Validate()
	if err != nil {
		return err
	}

	_, err = srv.relay.Save(cert)
	if err != nil {
		return err
	}

	srv.user.SetUserID(cert.TargetID)

	return err
}
