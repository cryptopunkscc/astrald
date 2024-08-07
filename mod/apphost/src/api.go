package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

func (mod *Module) regsiterApp(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	appID, ok := params["app_id"]
	if !ok {
		return astral.Reject()
	}

	id, err := mod.RegisterApp(appID)
	if err != nil {
		return astral.Reject()
	}

	return astral.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		id.WriteTo(w)
	})
}

func (mod *Module) unregsiterApp(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	appID, ok := params["app_id"]
	if !ok {
		return astral.Reject()
	}

	err := mod.UnregisterApp(appID)
	if err != nil {
		return astral.Reject()
	}

	return astral.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
	})
}

func (mod *Module) getAccessToken(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(q.Query)

	appID, ok := params["app_id"]
	if !ok {
		return astral.Reject()
	}

	token, err := mod.AppToken(appID)
	if err != nil {
		return astral.Reject()
	}

	return astral.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		cslq.Encode(conn, "[c]c", token)
	})
}

func (mod *Module) list(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return astral.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()
		cslq.Encode(conn, "[s][c]c", mod.ListApps())
	})
}
