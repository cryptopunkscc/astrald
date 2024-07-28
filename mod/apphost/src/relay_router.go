package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
)

type RelayRouter struct {
	log      *log.Logger
	target   string
	identity id.Identity
}

func (fwd *RelayRouter) RouteQuery(ctx context.Context, query astral.Query, caller io.WriteCloser, hints astral.Hints) (io.WriteCloser, error) {
	target, err := proto.Dial(fwd.target)
	if err != nil {
		fwd.log.Errorv(2, "%s:%s forward to %s: %s", query.Target(), query.Query(), fwd.target, err)
		return astral.Reject()
	}

	conn := proto.NewConn(target)

	err = conn.WriteMsg(proto.InQueryParams{
		Identity: query.Caller(),
		Query:    query.Query(),
	})
	if err != nil {
		target.Close()
		return astral.Reject()
	}

	if conn.ReadErr() != nil {
		target.Close()
		return astral.Reject()
	}

	go func() {
		io.Copy(caller, target)
		caller.Close()
	}()

	return target, nil
}
