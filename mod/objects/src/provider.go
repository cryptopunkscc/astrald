package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"slices"
)

type JSONDescriptor struct {
	Type string
	Data json.RawMessage
}

type Provider struct {
	mod    *Module
	router *core.PrefixRouter
}

func NewProvider(mod *Module) *Provider {
	var srv = &Provider{
		mod:    mod,
		router: core.NewPrefixRouter(true),
	}

	srv.router.EnableParams = true

	srv.router.AddRouteFunc(methodRead, srv.Read)
	srv.router.AddRouteFunc(methodDescribe, srv.Describe)
	srv.router.AddRouteFunc(methodPut, srv.Put)
	srv.router.AddRouteFunc(methodHold, srv.Hold)
	srv.router.AddRouteFunc(methodRelease, srv.Release)
	srv.router.AddRouteFunc(methodSearch, srv.Search)

	return srv
}

func (srv *Provider) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Read(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

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
	if err != nil && !errors.Is(err, core.ErrKeyNotFound) {
		srv.mod.log.Errorv(2, "offset: invalid argument: %v", err)
		return net.Reject()
	}

	r, err := srv.mod.OpenAs(ctx, query.Caller(), objectID, opts)
	if err != nil {
		srv.mod.log.Errorv(2, "open %v error: %v", objectID, err)
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.Conn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}

func (srv *Provider) Release(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.Conn) {
		defer conn.Close()

		srv.mod.Release(query.Caller(), objectID)
	})
}

func (srv *Provider) Hold(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionRead, objectID) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.Conn) {
		defer conn.Close()

		srv.mod.Hold(query.Caller(), objectID)
	})
}

func (srv *Provider) Describe(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionRead, objectID) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.Conn) {
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
	_, params := core.ParseQuery(query.Query())

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

	return net.Accept(query, caller, func(conn net.Conn) {
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

		srv.mod.Hold(query.Caller(), objectID)

		cslq.Encode(conn, "cv", 0, objectID)
	})
}

func (srv *Provider) Search(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !srv.mod.node.Auth().Authorize(query.Caller(), objects.ActionSearch) {
		return net.Reject()
	}

	_, params := core.ParseQuery(query.Query())

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

	return net.Accept(query, caller, func(conn net.Conn) {
		defer conn.Close()

		json.NewEncoder(conn).Encode(matches)

		return
	})
}
