package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"sync/atomic"
	"time"
)

type Query interface {
	Accept() (conn io.ReadWriteCloser, err error)
	Reject() (err error)
	Caller() *astral.Identity
	Extra() *sig.Map[string, any]
}

var _ Query = &NetworkQuery{}

type NetworkQuery struct {
	w io.WriteCloser
	*astral.Query
	r        chan queryResponse
	resolved atomic.Bool
}

func NewNetworkQuery(w io.WriteCloser, query *astral.Query) *NetworkQuery {
	return &NetworkQuery{
		w:     w,
		Query: query,
		r:     make(chan queryResponse, 1),
	}
}

func (e *NetworkQuery) Accept() (conn io.ReadWriteCloser, err error) {
	if !e.resolved.CompareAndSwap(false, true) {
		return nil, astral.NewError("already resolved")
	}

	var w io.WriteCloser
	var ch = make(chan io.ReadWriteCloser, 1)

	w, err = query.Accept(e.Query, e.w, func(c astral.Conn) {
		ch <- c
	})
	e.r <- queryResponse{w, err}
	if err != nil {
		return
	}

	return <-ch, nil
}

func (e *NetworkQuery) Reject() (err error) {
	if !e.resolved.CompareAndSwap(false, true) {
		return astral.NewError("already resolved")
	}

	_, err = query.Reject()
	e.r <- queryResponse{nil, err}

	return nil
}

func (e *NetworkQuery) Caller() *astral.Identity {
	return e.Query.Caller
}

func (e *NetworkQuery) Resolve() (io.WriteCloser, error) {
	select {
	case r := <-e.r:
		return r.WriteCloser, r.Error
	case <-time.After(5 * time.Second):
		e.Reject()
		return query.Reject()
	}
}

func (e *NetworkQuery) Extra() *sig.Map[string, any] {
	return &e.Query.Extra
}

type queryResponse struct {
	io.WriteCloser
	Error error
}

func AcceptStream(q Query) (stream *astral.Stream, err error) {
	var rw io.ReadWriteCloser
	rw, err = q.Accept()
	if err != nil {
		return
	}

	return astral.NewStream(rw, astral.ExtractBlueprints(rw)), err
}

func AcceptTerminal(q Query) (t *Terminal, err error) {
	var rw io.ReadWriteCloser
	rw, err = q.Accept()
	if err != nil {
		return
	}

	return NewTerminal(rw), err
}
