package route

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/route/proto"
	"github.com/cryptopunkscc/astrald/net"
	"io"
)

const RouteServiceName = "net.route"

var _ net.Router = &RouteService{}

type RouteService struct {
	*Module
}

func (service *RouteService) Run(ctx context.Context) error {
	_, err := service.node.Services().Register(ctx, service.node.Identity(), RouteServiceName, service)

	return err
}

func (service *RouteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		if err := service.serve(ctx, conn, query.Origin()); err != nil {
			service.log.Errorv(2, "(%s) serve: %s", query.Caller(), err)
		}
	})
}

func (service *RouteService) serve(ctx context.Context, conn net.SecureConn, origin string) error {
	var err error
	var caller = conn.RemoteIdentity()
	var c = proto.New(conn)
	defer c.Close()

	service.log.Logv(2, "(%s) connected", caller)

	for {
		var cmd proto.Cmd
		err = c.Decode(&cmd)
		if err != nil {
			return err
		}

		service.log.Logv(2, "(%s) cmd: %s", caller, cmd)

		switch cmd.Cmd {
		case proto.CmdCert:
			var cert proto.RelayCert
			if err := c.Decode(&cert); err != nil {
				return err
			}

			if !cert.Relay.IsEqual(caller) {
				return errors.New("certificate identity mismatch")
			}

			if err := cert.Verify(); err != nil {
				service.log.Logv(2, "(%s) route to %s denied: %v", caller, cert.Identity, err)
				if err := c.EncodeErr(proto.ErrDenied); err != nil {
					return err
				}
				continue
			}

			service.log.Logv(2, "(%s) shifted to %s", caller, cert.Identity)

			caller = cert.Identity

			if err := c.EncodeErr(nil); err != nil {
				return err
			}

		case proto.CmdQuery:
			var query proto.QueryParams
			if err := c.Decode(&query); err != nil {
				return err
			}

			srv, err := service.node.Services().Find(query.Target, query.Query)
			if err != nil {
				return c.EncodeErr(proto.ErrRejected)
			}
			if !srv.Identity().IsEqual(query.Target) {
				return c.EncodeErr(proto.ErrRejected)
			}

			var target = service.node.Identity()
			if !query.Target.IsEqual(target) {
				target, err = service.keys.Find(query.Target)
				if err != nil {
					return c.EncodeErr(proto.ErrUnableToProcess)
				}
			}

			c.EncodeErr(nil)

			if !target.IsEqual(service.node.Identity()) {
				var cert = proto.NewRelayCert(target, service.node.Identity())
				if err = cert.Sign(); err != nil {
					return err
				}

				if err = c.Encode(cert); err != nil {
					return err
				}
			}

			var q = net.NewQueryOrigin(caller, target, query.Query, origin)
			var shiftedConn = replaceIdentity{SecureConn: conn, remoteIdentity: caller}

			service.log.Logv(2, "(%s) routing query -> %s:%s", caller, q.Target(), q.Query())

			localWriter, err := srv.RouteQuery(ctx, q, shiftedConn)
			if err != nil {
				return c.EncodeErr(proto.ErrRejected)
			}

			if err := c.EncodeErr(nil); err != nil {
				return err
			}

			io.Copy(localWriter, conn)
			localWriter.Close()

			return nil

		default:
			return c.EncodeErr(proto.ErrInvalidRequest)
		}
	}
}
