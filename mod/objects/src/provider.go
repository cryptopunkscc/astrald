package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"slices"
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
	srv.router.AddRouteFunc(putServiceName, srv.Put)
	srv.router.AddRouteFunc(searchServiceName, srv.Search)

	return srv
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

	var opts = objects.DefaultOpenOpts()

	if hints.Origin == net.OriginLocal {
		opts.Zone |= net.ZoneNetwork
	}

	opts.Offset, err = params.GetUint64("offset")
	if err != nil && !errors.Is(err, router.ErrKeyNotFound) {
		srv.mod.log.Errorv(2, "offset: invalid argument: %v", err)
		return net.Reject()
	}

	r, err := srv.mod.OpenAs(ctx, query.Caller(), objectID, opts)
	if err != nil {
		srv.mod.log.Errorv(2, "open %v error: %v", objectID, err)
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

	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionRead, objectID) {
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

func (srv *Provider) Put(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionWrite) {
		return net.Reject()
	}

	size, err := params.GetUint64("size")
	if err != nil {
		return net.Reject()
	}

	w, err := srv.mod.Create(&objects.CreateOpts{Alloc: int(size)})
	if err != nil {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		_, err := io.CopyN(w, conn, int64(size))
		if err != nil {
			cslq.Encode(conn, "c", 1)
			return
		}

		objectID, err := w.Commit()
		if err != nil {
			cslq.Encode(conn, "c", 1)
			return
		}

		cslq.Encode(conn, "cv", 0, objectID)
	})
}

func (srv *Provider) Search(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionSearch) {
		return net.Reject()
	}

	_, params := router.ParseQuery(query.Query())

	q, ok := params["q"]
	if !ok {
		return net.Reject()
	}

	matches, err := srv.mod.Search(ctx, q, objects.DefaultSearchOpts())
	if err != nil {
		return net.Reject()
	}

	matches = slices.DeleteFunc(matches, func(match objects.Match) bool {
		return !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionRead, match.ObjectID)
	})

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		json.NewEncoder(conn).Encode(matches)

		return
	})
}
