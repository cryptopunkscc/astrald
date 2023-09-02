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

func (service *RouteService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		if err := service.serve(ctx, conn, hints.Origin); err != nil {
			service.log.Errorv(2, "(%s) serve: %s", query.Caller(), err)
		}
	})
}

func (service *RouteService) serve(ctx context.Context, conn net.SecureConn, origin string) error {
	var err error
	var caller = conn.RemoteIdentity()
	var rpc = proto.New(conn)
	defer rpc.Close()

	for {
		var cmd proto.Cmd
		err = rpc.Decode(&cmd)
		if err != nil {
			return err
		}

		switch cmd.Cmd {
		case proto.CmdCert:
			var cert proto.RelayCert
			if err := rpc.Decode(&cert); err != nil {
				return err
			}

			if !cert.Relay.IsEqual(caller) {
				return errors.New("certificate identity mismatch")
			}

			if err := cert.Verify(); err != nil {
				service.log.Logv(2, "%s provided an invalid certificate for %s: %v", caller, cert.Identity, err)
				if err := rpc.EncodeErr(proto.ErrDenied); err != nil {
					return err
				}
				continue
			}

			service.log.Logv(2, "%s shifted to %s", caller, cert.Identity)

			caller = cert.Identity

			if err := rpc.EncodeErr(nil); err != nil {
				return err
			}

		case proto.CmdQuery:
			var params proto.QueryParams
			if err := rpc.Decode(&params); err != nil {
				return err
			}

			var target = service.node.Identity()

			// if query target is not the node, look up private keys for the target identity
			if !params.Target.IsEqual(target) {
				target, err = service.keys.Find(params.Target)
				if err != nil {
					return rpc.EncodeErr(proto.ErrUnableToProcess)
				}
			}

			rpc.EncodeErr(nil)

			// send a certificate if node is not the target
			if !target.IsEqual(service.node.Identity()) {
				var cert = proto.NewRelayCert(target, service.node.Identity())
				if err = cert.Sign(); err != nil {
					return err
				}

				if err = rpc.Encode(cert); err != nil {
					return err
				}
			}

			var query = net.NewQuery(caller, target, params.Query)
			var shiftedConn = &replaceIdentity{SecureConn: conn, remoteIdentity: caller}
			shiftedConn.Lock()

			localWriter, err := service.node.Router().RouteQuery(ctx, query, shiftedConn, net.Hints{Origin: origin})
			if err != nil {
				return rpc.EncodeErr(proto.ErrRejected)
			}

			if err := rpc.EncodeErr(nil); err != nil {
				return err
			}

			shiftedConn.Unlock()
			io.Copy(localWriter, conn)
			localWriter.Close()

			return nil

		default:
			return rpc.EncodeErr(proto.ErrInvalidRequest)
		}
	}
}
