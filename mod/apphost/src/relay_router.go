package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
)

type RelayRouter struct {
	log      *log.Logger
	target   string
	identity *astral.Identity
}

func (fwd *RelayRouter) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	target, err := proto.Dial(fwd.target)
	if err != nil {
		fwd.log.Errorv(2, "%s:%s forward to %s: %s", q.Target, q.Query, fwd.target, err)
		return query.Reject()
	}

	conn := proto.NewConn(target)

	err = conn.WriteMsg(proto.InQueryParams{
		Identity: q.Caller,
		Query:    q.Query,
	})
	if err != nil {
		target.Close()
		return query.Reject()
	}

	if conn.ReadErr() != nil {
		target.Close()
		return query.Reject()
	}

	go func() {
		io.Copy(w, target)
		w.Close()
	}()

	return target, nil
}
