package objects

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"slices"
	"strconv"
)

type JSONDescriptor struct {
	Type string
	Data json.RawMessage
}

type Provider struct {
	mod    *Module
	router *router.PrefixRouter
}

func NewProvider(mod *Module) *Provider {
	var srv = &Provider{
		mod:    mod,
		router: router.NewPrefixRouter(true),
	}

	srv.router.EnableParams = true

	srv.router.AddRouteFunc(readServiceName, srv.Read)
	srv.router.AddRouteFunc(describeServiceName, srv.Describe)

	return srv
}

func (srv *Provider) Run(ctx context.Context) error {
	return nil
}

func (srv *Provider) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Read(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ReadAction, objectID) {
		return net.Reject()
	}

	var opts = &objects.OpenOpts{}
	if s, found := params["offset"]; found {
		opts.Offset, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			srv.mod.log.Errorv(2, "parse offset error: %v", err)
			return net.Reject()
		}
	}

	r, err := srv.mod.Open(context.Background(), objectID, opts)
	if err != nil {
		srv.mod.log.Errorv(2, "read %v error: %v", objectID, err)
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}

func (srv *Provider) Describe(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	if !srv.mod.node.Auth().Authorize(query.Caller(), shares.DescribeAction, objectID) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		var list []JSONDescriptor

		for _, d := range srv.mod.Describe(ctx, objectID, nil) {
			if !slices.Contains(srv.mod.config.DescriptorWhitelist, d.Data.Type()) {
				continue
			}

			b, err := json.Marshal(d.Data)
			if err != nil {
				continue
			}

			list = append(list, JSONDescriptor{
				Type: d.Data.Type(),
				Data: b,
			})
		}

		json.NewEncoder(conn).Encode(list)
	})
}
