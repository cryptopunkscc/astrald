package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"github.com/cryptopunkscc/astrald/net"
	"io"
)

type RelayRouter struct {
	log      *log.Logger
	target   string
	identity id.Identity
}

func (fwd *RelayRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	target, err := proto.Dial(fwd.target)
	if err != nil {
		fwd.log.Errorv(2, "%s:%s forward to %s: %s", query.Target(), query.Query(), fwd.target, err)
		return net.Reject()
	}

	conn := proto.NewConn(target)

	err = conn.WriteMsg(proto.InQueryParams{
		Identity: query.Caller(),
		Query:    query.Query(),
	})
	if err != nil {
		target.Close()
		return net.Reject()
	}

	if conn.ReadErr() != nil {
		target.Close()
		return net.Reject()
	}

	go func() {
		io.Copy(caller, target)
		caller.Close()
	}()

	return net.NewSecurePipeWriter(target, query.Target()), nil
}
