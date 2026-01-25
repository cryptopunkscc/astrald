package ops

import (
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/sig"
)

type Query struct {
	w io.WriteCloser
	*astral.Query
	r        chan queryResponse
	resolved atomic.Bool
}

func newQuery(w io.WriteCloser, query *astral.Query) *Query {
	return &Query{
		w:     w,
		Query: query,
		r:     make(chan queryResponse, 1),
	}
}

func (e *Query) Accept() (conn io.ReadWriteCloser) {
	if !e.resolved.CompareAndSwap(false, true) {
		return ErrorConn{Err: errors.New("query already resolved")}
	}

	var w io.WriteCloser
	var ch = make(chan io.ReadWriteCloser, 1)

	w, err := query.Accept(e.Query, e.w, func(c astral.Conn) {
		ch <- c
	})
	e.r <- queryResponse{w, err}
	if err != nil {
		return ErrorConn{Err: err}
	}

	return <-ch
}

func (e *Query) AcceptChannel(cfg ...channel.ConfigFunc) *channel.Channel {
	return channel.New(e.Accept(), cfg...)
}

func (e *Query) Reject() (err error) {
	if !e.resolved.CompareAndSwap(false, true) {
		return errors.New("query already resolved")
	}

	_, err = query.Reject()
	e.r <- queryResponse{nil, err}

	return nil
}

func (e *Query) RejectWithCode(code uint8) (err error) {
	if !e.resolved.CompareAndSwap(false, true) {
		return errors.New("query already resolved")
	}

	if code == 0 {
		code = 1
	}

	_, err = query.RejectWithCode(code)
	e.r <- queryResponse{nil, err}

	return nil
}

func (e *Query) Caller() *astral.Identity {
	return e.Query.Caller
}

func (e *Query) Extra() *sig.Map[string, any] {
	return &e.Query.Extra
}

func (e *Query) Origin() string {
	if v, ok := e.Extra().Get("origin"); ok && v != nil {
		return v.(string)
	}
	return ""
}

func (e *Query) resolve(ctx *astral.Context) (io.WriteCloser, error) {
	select {
	case r := <-e.r:
		return r.WriteCloser, r.Error

	case <-ctx.Done():
		e.Reject()
		return query.RejectWithCode(astral.CodeCanceled)

	case <-time.After(5 * time.Second):
		e.Reject()
		return query.Reject()
	}
}

type queryResponse struct {
	io.WriteCloser
	Error error
}
