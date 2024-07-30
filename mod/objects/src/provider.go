package objects

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"slices"
)

const maxPushSize = 4096

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
	srv.router.AddRouteFunc(methodSearch, srv.Search)
	srv.router.AddRouteFunc(methodPush, srv.Push)

	return srv
}

func (srv *Provider) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller)
}

func (srv *Provider) Read(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(query.Query)

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	var opts = objects.DefaultOpenOpts()

	if o, ok := query.Extra.Get("origin"); ok && (o == astral.OriginLocal) {
		opts.Zone |= astral.ZoneNetwork
	}

	opts.Offset, err = params.GetUint64("offset")
	if err != nil && !errors.Is(err, core.ErrKeyNotFound) {
		srv.mod.log.Errorv(2, "offset: invalid argument: %v", err)
		return astral.Reject()
	}

	r, err := srv.mod.OpenAs(ctx, query.Caller, objectID, opts)
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

func (srv *Provider) Describe(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(query.Query)

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	if !srv.mod.Auth.Authorize(query.Caller, objects.ActionRead, objectID) {
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

func (srv *Provider) Put(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(query.Query)

	if !srv.mod.Auth.Authorize(query.Caller, objects.ActionWrite) {
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

		cslq.Encode(conn, "cv", 0, objectID)
	})
}

func (srv *Provider) Search(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	if !srv.mod.Auth.Authorize(query.Caller, objects.ActionSearch) {
		return astral.Reject()
	}

	_, params := core.ParseQuery(query.Query)

	q, ok := params["q"]
	if !ok {
		return astral.Reject()
	}

	matches, err := srv.mod.Search(ctx, q, objects.DefaultSearchOpts())
	if err != nil {
		return astral.Reject()
	}

	matches = slices.DeleteFunc(matches, func(match objects.Match) bool {
		return !srv.mod.Auth.Authorize(query.Caller, objects.ActionRead, match.ObjectID)
	})

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		json.NewEncoder(conn).Encode(matches)

		return
	})
}

func (srv *Provider) Push(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(query.Query)

	size, err := params.GetInt("size")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return astral.Reject()
	}

	if size > maxPushSize {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		var buf = make([]byte, size)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			srv.mod.log.Errorv(1, "%v push read error: %v", query.Caller, err)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		obj, err := srv.mod.ReadObject(bytes.NewReader(buf))
		if err != nil {
			srv.mod.log.Errorv(1, "%v push read object error: %v", query.Caller, err)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		var push = &objects.Push{
			Source:   query.Caller,
			ObjectID: object.Resolve(buf),
			Object:   obj,
		}

		if !srv.mod.pushLocal(push) {
			srv.mod.log.Errorv(1, "rejected %s from %v (%v)", obj.ObjectType(), query.Caller, push.ObjectID)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		binary.Write(conn, binary.BigEndian, true)

		srv.mod.log.Infov(1, "received %s from %v (%v)", obj.ObjectType(), query.Caller, push.ObjectID)

		return
	})
}
