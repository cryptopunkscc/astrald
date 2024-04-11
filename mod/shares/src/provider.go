package shares

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"slices"
	"strconv"
	"time"
)

const (
	opDone     = 0x00
	opAdd      = 0x01
	opRemove   = 0x02
	opResync   = 0x03
	opNotFound = 0xff
)

const readServiceName = "shares.read"
const notifyServiceName = "shares.notify"
const describeServiceName = "shares.describe"

type JSONDescriptor struct {
	Type string
	Data json.RawMessage
}

type Provider struct {
	*Module
	router *router.PrefixRouter
}

func (srv *Provider) Run(ctx context.Context) error {
	return nil
}

func NewProvider(mod *Module) *Provider {
	var srv = &Provider{
		Module: mod,
		router: router.NewPrefixRouter(true),
	}

	srv.router.EnableParams = true

	srv.router.AddRouteFunc("shares.read", srv.Read)
	srv.router.AddRouteFunc("shares.sync", srv.Sync)
	srv.router.AddRouteFunc("shares.describe", srv.Describe)
	srv.router.AddRouteFunc("shares.notify", srv.Notify)

	return srv
}

func (srv *Provider) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Read(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	dataID, err := params.GetDataID("id")
	if err != nil {
		srv.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	if !srv.node.Auth().Authorize(query.Caller(), storage.OpenAction, dataID) {
		return net.Reject()
	}

	var opts = &storage.OpenOpts{Virtual: true}
	if s, found := params["offset"]; found {
		opts.Offset, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			srv.log.Errorv(2, "parse offset error: %v", err)
			return net.Reject()
		}
	}

	r, err := srv.storage.Open(dataID, opts)
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

func (srv *Provider) Sync(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var since = time.Time{}

	_, params := router.ParseQuery(query.Query())
	ts, _ := params.GetInt("since")
	if ts > 0 {
		since = time.Unix(0, int64(ts))
	}

	if since.After(time.Now()) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		var before = time.Now()
		var updateMode = !since.IsZero()

		share, err := srv.FindLocalShare(caller.Identity())
		if err != nil {
			cslq.Encode(conn, "c", opNotFound)
			return
		}

		if updateMode && share.TrimmedAt().After(since) {
			cslq.Encode(conn, "c", opResync)
			return
		}

		entries, err := share.Scan(&sets.ScanOpts{
			UpdatedAfter:   since,
			UpdatedBefore:  before,
			IncludeRemoved: updateMode,
		})
		if err != nil {
			return
		}

		for _, entry := range entries {
			var op byte
			if entry.Removed {
				op = opRemove
			} else {
				op = opAdd
			}

			err = cslq.Encode(conn, "cv",
				op,
				entry.DataID,
			)

			if err != nil {
				return
			}
		}

		cslq.Encode(conn, "cq", opDone, before.UnixNano())
	})
}

func (srv *Provider) Describe(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	dataID, err := params.GetDataID("id")
	if err != nil {
		srv.log.Errorv(2, "invalid id: %v", err)
		return net.Reject()
	}

	if !srv.node.Auth().Authorize(query.Caller(), shares.DescribeAction, dataID) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		var list []JSONDescriptor

		for _, d := range srv.content.Describe(ctx, dataID, nil) {
			if !slices.Contains(srv.config.DescriptorWhitelist, d.Data.Type()) {
				continue
			}

			b, err := json.Marshal(d.Data)
			if err != nil {
				continue
			}

			list = append(list, JSONDescriptor{
				Type: d.Data.Type(),
				Data: json.RawMessage(b),
			})
		}

		json.NewEncoder(conn).Encode(list)
	})
}

func (srv *Provider) Notify(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	remoteShare, err := srv.FindRemoteShare(query.Target(), query.Caller())
	if err != nil {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		conn.Close()
		srv.tasks <- func(ctx context.Context) {
			remoteShare.Sync(ctx)
		}
	})
}
