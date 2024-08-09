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
	"github.com/cryptopunkscc/astrald/lib/query"
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

func (srv *Provider) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return srv.router.RouteQuery(ctx, q, w)
}

func (srv *Provider) Read(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return query.Reject()
	}

	var opts = objects.DefaultOpenOpts()

	if o, ok := q.Extra.Get("origin"); ok && (o == astral.OriginLocal) {
		opts.Zone |= astral.ZoneNetwork
	}

	opts.Offset, err = params.GetUint64("offset")
	if err != nil && !errors.Is(err, core.ErrKeyNotFound) {
		srv.mod.log.Errorv(2, "offset: invalid argument: %v", err)
		return query.Reject()
	}

	r, err := srv.mod.OpenAs(ctx, q.Caller, objectID, opts)
	if err != nil {
		srv.mod.log.Errorv(2, "open %v error: %v", objectID, err)
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}

func (srv *Provider) Describe(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	objectID, err := params.GetObjectID("id")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return query.Reject()
	}

	if !srv.mod.Auth.Authorize(q.Caller, objects.ActionRead, &objectID) {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
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

func (srv *Provider) Put(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	if !srv.mod.Auth.Authorize(q.Caller, objects.ActionWrite, nil) {
		return query.Reject()
	}

	size, err := params.GetUint64("size")
	if err != nil {
		return query.Reject()
	}

	create, err := srv.mod.Create(&objects.CreateOpts{Alloc: int(size)})
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		_, err := io.CopyN(create, conn, int64(size))
		if err != nil {
			cslq.Encode(conn, "c", 1)
			return
		}

		objectID, err := create.Commit()
		if err != nil {
			cslq.Encode(conn, "c", 1)
			return
		}

		cslq.Encode(conn, "cv", 0, objectID)
	})
}

func (srv *Provider) Search(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	if !srv.mod.Auth.Authorize(q.Caller, objects.ActionSearch, nil) {
		return query.Reject()
	}

	_, params := core.ParseQuery(q.Query)

	sq, ok := params["q"]
	if !ok {
		return query.Reject()
	}

	matches, err := srv.mod.Search(ctx, sq, objects.DefaultSearchOpts())
	if err != nil {
		return query.Reject()
	}

	matches = slices.DeleteFunc(matches, func(match objects.Match) bool {
		return !srv.mod.Auth.Authorize(q.Caller, objects.ActionRead, &match.ObjectID)
	})

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		json.NewEncoder(conn).Encode(matches)

		return
	})
}

func (srv *Provider) Push(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	size, err := params.GetInt("size")
	if err != nil {
		srv.mod.log.Errorv(2, "invalid id: %v", err)
		return query.Reject()
	}

	if size > maxPushSize {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		var buf = make([]byte, size)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			srv.mod.log.Errorv(1, "%v push read error: %v", q.Caller, err)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		obj, err := srv.mod.ReadObject(bytes.NewReader(buf))
		if err != nil {
			srv.mod.log.Errorv(1, "%v push read object error: %v", q.Caller, err)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		var objectID object.ID
		objectID, err = astral.ResolveObjectID(obj)
		if err != nil {
			return
		}

		var push = &objects.SourcedObject{
			Source: q.Caller,
			Object: obj,
		}

		if !srv.mod.pushLocal(push) {
			srv.mod.log.Errorv(1, "rejected %s from %v (%v)", obj.ObjectType(), q.Caller, objectID)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		binary.Write(conn, binary.BigEndian, true)

		srv.mod.log.Infov(1, "accepted %s from %v (%v)", obj.ObjectType(), q.Caller, objectID)

		return
	})
}
