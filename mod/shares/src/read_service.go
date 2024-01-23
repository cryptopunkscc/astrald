package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"strings"
)

const readServicePrefix = "shares.read."

type ReadService struct {
	*Module
}

func NewReadService(module *Module) *ReadService {
	return &ReadService{Module: module}
}

func (srv *ReadService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(readServicePrefix+"*", srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(readServicePrefix + "*")

	<-ctx.Done()
	return nil
}

func (srv *ReadService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	idstr, found := strings.CutPrefix(query.Query(), readServicePrefix)
	if !found {
		return net.Reject()
	}

	dataID, err := data.Parse(idstr)
	if err != nil {
		srv.log.Errorv(2, "parse error: %v", err)
		return net.Reject()
	}

	err = srv.Authorize(query.Caller(), dataID)
	if err != nil {
		srv.log.Errorv(2, "access to %v denied for %v (%v)", dataID, query.Caller(), err)
		return net.Reject()
	}

	r, err := srv.storage.Read(dataID, nil)
	if err != nil {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}
