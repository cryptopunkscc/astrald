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
	"strings"
	"time"
)

const maxPushSize = 4096

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

func (p *Provider) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return p.router.RouteQuery(ctx, q, w)
}

func (p *Provider) Read(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	objectID, err := params.GetObjectID("id")
	if err != nil {
		p.mod.log.Errorv(2, "invalid id: %v", err)
		return query.Reject()
	}

	var opts = objects.DefaultOpenOpts()

	if q.IsLocal() {
		opts.Zone |= astral.ZoneNetwork
	}

	opts.Offset, err = params.GetUint64("offset")
	if err != nil && !errors.Is(err, core.ErrKeyNotFound) {
		p.mod.log.Errorv(2, "offset: invalid argument: %v", err)
		return query.Reject()
	}

	r, err := p.mod.OpenAs(ctx, q.Caller, objectID, opts)
	if err != nil {
		p.mod.log.Errorv(2, "open %v error: %v", objectID, err)
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer r.Close()
		defer conn.Close()

		io.Copy(conn, r)
	})
}

func (p *Provider) Put(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	if !p.mod.Auth.Authorize(q.Caller, objects.ActionWrite, nil) {
		return query.Reject()
	}

	size, err := params.GetUint64("size")
	if err != nil {
		return query.Reject()
	}

	create, err := p.mod.Create(&objects.CreateOpts{Alloc: int(size)})
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

func (p *Provider) Search(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	var args struct {
		Query  string `query:"key:q"`
		Zones  string `query:"optional"`
		Format string `query:"optional"`
		Ext    string `query:"optional"`
	}
	_, err := query.ParseTo(q.Query, &args)
	if err != nil {
		p.mod.log.Errorv(2, "Search(): parse args: %v", err)
		return query.Reject()
	}

	opts := objects.DefaultSearchOpts()
	opts.ClientID = q.Caller

	// only local queries can modify the scope of the search
	if q.IsLocal() {
		if len(args.Zones) > 0 {
			opts.Zone = astral.Zones(args.Zones)
		}

		if len(args.Ext) > 0 {
			var ids []*astral.Identity
			targets := strings.Split(args.Ext, ",")
			for _, target := range targets {
				id, err := p.mod.Dir.Resolve(target)
				if err != nil {
					return query.Reject()
				}
				ids = append(ids, id)
			}
			opts.Extra.Set("ext", ids)
		}
	}

	sctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	matches, err := p.mod.Search(sctx, args.Query, opts)
	if err != nil {
		cancel()
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		defer cancel()

		var dup = make(map[string]struct{})

		for match := range matches {
			if !p.mod.Auth.Authorize(q.Caller, objects.ActionRead, &match.ObjectID) {
				continue
			}

			if _, found := dup[match.ObjectID.String()]; found {
				continue
			}

			dup[match.ObjectID.String()] = struct{}{}

			switch args.Format {
			case "json":
				err = json.NewEncoder(conn).Encode(match)
			default:
				_, err = match.WriteTo(conn)
			}

			if err != nil {
				return
			}
		}
	})
}

func (p *Provider) Push(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	size, err := params.GetInt("size")
	if err != nil {
		p.mod.log.Errorv(2, "invalid id: %v", err)
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
			p.mod.log.Errorv(1, "%v push read error: %v", q.Caller, err)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		obj, err := p.mod.ReadObject(bytes.NewReader(buf))
		if err != nil {
			p.mod.log.Errorv(1, "%v push read object error: %v", q.Caller, err)
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

		if !p.mod.pushLocal(push) {
			p.mod.log.Errorv(1, "rejected %s from %v (%v)", obj.ObjectType(), q.Caller, objectID)
			binary.Write(conn, binary.BigEndian, false)
			return
		}

		binary.Write(conn, binary.BigEndian, true)

		p.mod.log.Infov(1, "accepted %s from %v (%v)", obj.ObjectType(), q.Caller, objectID)

		return
	})
}
