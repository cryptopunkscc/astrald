package router

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/router/proto"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const RouterServiceName = "net.router"

var _ net.Router = &RouterService{}

type RouterService struct {
	*Module
}

func (srv *RouterService) Run(ctx context.Context) error {
	_, err := srv.node.Services().Register(ctx, srv.node.Identity(), RouterServiceName, srv)

	return err
}

func (srv *RouterService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		if err := srv.serve(ctx, conn); err != nil {
			srv.log.Errorv(2, "error serving %s: %s", query.Caller(), err)
		}
	})
}

func (srv *RouterService) serve(ctx context.Context, conn net.SecureConn) error {
	defer conn.Close()

	var err error
	var callerIM = NewIdentityMachine(conn.RemoteIdentity())
	var session = proto.New(conn)

	// get query params
	var params proto.QueryParams
	if err = session.Decode(&params); err != nil {
		return err
	}

	// apply caller certificate
	if len(params.Cert) > 0 {
		if err = callerIM.Apply(params.Cert); err != nil {
			session.EncodeErr(proto.ErrUnableToProcess)
			return err
		}
	}

	var response proto.QueryResponse

	// attach target certificate if necessary
	if !params.Target.IsEqual(srv.node.Identity()) {
		// look up private keys for the target identity
		targetKey, err := srv.keys.Find(params.Target)
		if err != nil {
			_ = session.EncodeErr(proto.ErrUnableToProcess)
			return errors.New("private key for target identity missing")
		}

		// create a relay certificate
		var cert = NewRouterCert(targetKey, srv.node.Identity(), time.Now().Add(time.Minute))
		response.Cert, err = cslq.Marshal(cert)
		if err != nil {
			_ = session.EncodeErr(proto.ErrUnableToProcess)
			return err
		}
	}

	// create a proxy service
	redirectCtx, _ := context.WithTimeout(ctx, time.Minute)
	var realQuery = net.NewQuery(callerIM.identity, params.Target, params.Query)

	redirect, err := NewRedirect(redirectCtx, realQuery, conn.RemoteIdentity(), srv.node)
	if err != nil {
		session.EncodeErr(proto.ErrUnableToProcess)
		return err
	}

	response.ProxyService = redirect.ServiceName

	// send response
	_ = session.EncodeErr(nil)
	return session.Encode(response)
}
