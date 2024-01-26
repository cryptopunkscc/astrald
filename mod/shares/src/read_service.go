package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"strconv"
)

const readServiceName = "shares.read"

type ReadService struct {
	*Module
}

func NewReadService(module *Module) *ReadService {
	return &ReadService{Module: module}
}

func (srv *ReadService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(readServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(readServiceName)

	<-ctx.Done()
	return nil
}

func (srv *ReadService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	idstr, found := params["id"]
	if !found {
		srv.log.Errorv(2, "query from %v contains no id parameter", query.Caller())
		return net.Reject()
	}

	dataID, err := data.Parse(idstr)
	if err != nil {
		srv.log.Errorv(2, "parse id error: %v", err)
		return net.Reject()
	}

	err = srv.Authorize(query.Caller(), dataID)
	if err != nil {
		srv.log.Errorv(2, "access to %v denied for %v (%v)", dataID, query.Caller(), err)
		return net.Reject()
	}

	var opts = &storage.ReadOpts{Virtual: true}
	if s, found := params["offset"]; found {
		opts.Offset, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			srv.log.Errorv(2, "parse offset error: %v", err)
			return net.Reject()
		}
	}

	r, err := srv.storage.Read(dataID, opts)
	if err != nil {
		srv.log.Errorv(2, "read %v error: %v", dataID, err)
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}
