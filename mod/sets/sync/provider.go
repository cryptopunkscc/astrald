package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"io"
	"time"
)

type Provider struct {
	set sets.Set
}

func NewProvider(set sets.Set) *Provider {
	return &Provider{set: set}
}

func (srv *Provider) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	_, params := core.ParseQuery(query.Query)

	since, _ := params.GetUnixNano("since")

	if since.After(time.Now()) {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		var before = time.Now()
		var updateMode = !since.IsZero()

		if updateMode && srv.set.TrimmedAt().After(since) {
			cslq.Encode(conn, "c", opResync)
			return
		}

		entries, err := srv.set.Scan(&sets.ScanOpts{
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
				entry.ObjectID,
			)

			if err != nil {
				return
			}
		}

		cslq.Encode(conn, "cq", opDone, before.UnixNano())
	})
}
