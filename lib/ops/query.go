package ops

import (
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/sig"
)

type Query struct {
	*astral.Query
	remoteWriter io.WriteCloser
	response     chan queryResponse
	resolved     atomic.Bool
}

func newQuery(remoteWriter io.WriteCloser, query *astral.Query) *Query {
	return &Query{
		remoteWriter: remoteWriter,
		Query:        query,
		response:     make(chan queryResponse, 1),
	}
}

func (query *Query) Accept() (conn io.ReadWriteCloser) {
	if !query.resolved.CompareAndSwap(false, true) {
		return ErrorConn{Err: errors.New("query already resolved")}
	}

	localReader, localWriter := io.Pipe()

	conn = NewConn(query.Query.Caller, query.Query.Target, query.remoteWriter, localReader, false)

	query.response <- queryResponse{localWriter, nil}

	return conn
}

func (query *Query) AcceptChannel(cfg ...channel.ConfigFunc) *channel.Channel {
	return channel.New(query.Accept(), cfg...)
}

func (query *Query) Reject() (err error) {
	if !query.resolved.CompareAndSwap(false, true) {
		return errors.New("query already resolved")
	}

	err = &astral.ErrRejected{Code: 0}
	query.response <- queryResponse{nil, err}

	return nil
}

func (query *Query) RejectWithCode(code uint8) (err error) {
	if !query.resolved.CompareAndSwap(false, true) {
		return errors.New("query already resolved")
	}

	if code == 0 {
		code = 1
	}

	err = astral.NewErrRejected(code)
	query.response <- queryResponse{nil, err}

	return nil
}

func (query *Query) Caller() *astral.Identity {
	return query.Query.Caller
}

func (query *Query) Extra() *sig.Map[string, any] {
	return &query.Query.Extra
}

func (query *Query) Origin() string {
	if v, ok := query.Extra().Get("origin"); ok && v != nil {
		return v.(string)
	}
	return ""
}

func (query *Query) resolve(ctx *astral.Context) (io.WriteCloser, error) {
	select {
	case r := <-query.response:
		return r.WriteCloser, r.Error

	case <-ctx.Done():
		query.Reject()
		return nil, astral.NewErrRejected(astral.CodeCanceled)

	case <-time.After(5 * time.Second):
		query.Reject()
		return nil, astral.NewErrRejected(0)
	}
}

type queryResponse struct {
	io.WriteCloser
	Error error
}
