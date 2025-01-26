package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"io"
)

type Relay struct {
	token    string
	log      *log.Logger
	endpoint string
}

func NewRelay(endpoint string, token string, log *log.Logger) *Relay {
	return &Relay{
		endpoint: endpoint,
		token:    token,
		log:      log,
	}
}

func (fwd *Relay) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	conn, err := ipc.Dial(fwd.endpoint)
	if err != nil {
		if fwd.log != nil {
			fwd.log.Errorv(2, "%v:%v forward to %v: %v", q.Target, q.Query, fwd.endpoint, err)
		}
		return query.Reject()
	}

	_, err = apphost.QueryInfo{
		Token:  astral.String8(fwd.token),
		Caller: q.Caller,
		Target: q.Target,
		Query:  astral.String16(q.Query),
	}.WriteTo(conn)
	if err != nil {
		conn.Close()
		return query.Reject()
	}

	var res astral.Uint8
	_, err = res.ReadFrom(conn)
	if err != nil {
		return query.RouteNotFound(fwd)
	}
	if res > 0 {
		return query.RejectWithCode(uint8(res))
	}

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()

	return conn, nil
}
