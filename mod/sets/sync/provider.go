package sync

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"time"
)

type Provider struct {
	set sets.Set
}

func NewProvider(set sets.Set) *Provider {
	return &Provider{set: set}
}

func (srv *Provider) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	_, params := router.ParseQuery(query.Query())

	since, _ := params.GetUnixNano("since")

	if since.After(time.Now()) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
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
				entry.DataID,
			)

			if err != nil {
				return
			}
		}

		cslq.Encode(conn, "cq", opDone, before.UnixNano())
	})
}
