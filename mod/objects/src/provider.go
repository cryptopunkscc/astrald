package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"slices"
)

type JSONDescriptor struct {
	Type string
	Data json.RawMessage
}

type Provider struct {
	mod    *Module
	router *routers.PrefixRouter
}

func NewProvider(mod *Module) *Provider {
	var srv = &Provider{
		mod:    mod,
		router: routers.NewPrefixRouter(true),
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

func (srv *Provider) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Read(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	var opts = objects.DefaultOpenOpts()

	if hints.Origin == astral.OriginLocal {
		opts.Zone |= astral.ZoneNetwork
	}

	opts.Offset, err = params.GetUint64("offset")
	if err != nil && !errors.Is(err, core.ErrKeyNotFound) {
		srv.mod.log.Errorv(2, "offset: invalid argument: %v", err)
		return astral.Reject()
	}

	r, err := srv.mod.OpenAs(ctx, query.Caller(), objectID, opts)
	if err != nil {
		srv.mod.log.Errorv(2, "open %v error: %v", objectID, err)
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}

func (srv *Provider) Release(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		srv.mod.Release(query.Caller(), objectID)
	})
}

func (srv *Provider) Hold(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	if !srv.mod.auth.Authorize(query.Caller(), objects.ActionRead, objectID) {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		srv.mod.Hold(query.Caller(), objectID)
	})
}

func (srv *Provider) Describe(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	if !srv.mod.auth.Authorize(query.Caller(), objects.ActionRead, objectID) {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
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

func (srv *Provider) Put(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	_, params := core.ParseQuery(query.Query())

	if !srv.mod.auth.Authorize(query.Caller(), objects.ActionWrite) {
		return astral.Reject()
	}

	size, err := params.GetUint64("size")
	if err != nil {
		return astral.Reject()
	}

	w, err := srv.mod.Create(&objects.CreateOpts{Alloc: int(size)})
	if err != nil {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
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

func (srv *Provider) Search(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	if !srv.mod.auth.Authorize(query.Caller(), objects.ActionSearch) {
		return astral.Reject()
	}

	_, params := core.ParseQuery(query.Query())

	q, ok := params["q"]
	if !ok {
		return astral.Reject()
	}

	matches, err := srv.mod.Search(ctx, q, objects.DefaultSearchOpts())
	if err != nil {
		return astral.Reject()
	}

	matches = slices.DeleteFunc(matches, func(match objects.Match) bool {
		return !srv.mod.auth.Authorize(query.Caller(), objects.ActionRead, match.ObjectID)
	})

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		json.NewEncoder(conn).Encode(matches)

		return
	})
}
