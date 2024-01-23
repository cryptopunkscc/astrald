package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"strconv"
	"strings"
	"time"
)

const syncServicePrefix = "shares.sync."

type SyncService struct {
	*Module
}

func NewSyncService(module *Module) *SyncService {
	return &SyncService{Module: module}
}

func (srv *SyncService) Run(ctx context.Context) error {
	err := srv.node.LocalRouter().AddRoute(syncServicePrefix+"*", srv)
	if err != nil {
		return err
	}
	defer srv.node.LocalRouter().RemoveRoute(syncServicePrefix + "*")

	<-ctx.Done()
	return nil
}

func (srv *SyncService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	arg, found := strings.CutPrefix(query.Query(), syncServicePrefix)
	if !found {
		return net.Reject()
	}

	ts, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return net.Reject()
	}

	var since = time.Time{}
	if ts > 0 {
		since = time.Unix(0, ts)
	}

	if since.After(time.Now()) {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		var indexName = srv.localShareIndexName(caller.Identity())

		entries, err := srv.index.UpdatedSince(indexName, since)
		if err != nil {
			return
		}

		for _, entry := range entries {
			var added byte
			if entry.Added {
				added = 1
			}

			err = cslq.Encode(conn, "vcv",
				cslq.Time(entry.UpdatedAt),
				added,
				entry.DataID,
			)

			if err != nil {
				return
			}
		}
	})
}
