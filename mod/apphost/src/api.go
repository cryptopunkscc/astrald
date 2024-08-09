package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
)

func (mod *Module) regsiterApp(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	appID, ok := params["app_id"]
	if !ok {
		return query.Reject()
	}

	id, err := mod.RegisterApp(appID)
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		id.WriteTo(w)
	})
}

func (mod *Module) unregsiterApp(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	appID, ok := params["app_id"]
	if !ok {
		return query.Reject()
	}

	err := mod.UnregisterApp(appID)
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
	})
}

func (mod *Module) getAccessToken(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	appID, ok := params["app_id"]
	if !ok {
		return query.Reject()
	}

	token, err := mod.AppToken(appID)
	if err != nil {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		cslq.Encode(conn, "[c]c", token)
	})
}

func (mod *Module) list(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		cslq.Encode(conn, "[s][c]c", mod.ListApps())
	})
}
