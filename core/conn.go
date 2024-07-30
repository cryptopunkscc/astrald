package core

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"sync/atomic"
)

type conn struct {
	router *Router
	query  *astral.Query
	src    *writer
	dst    *writer
}

func newConn(r *Router, q *astral.Query) *conn {
	return &conn{
		router: r,
		query:  q,
	}
}

func (c *conn) Close() error {
	c.src.Close()
	c.dst.Close()
	c.router.conns.Delete(c.query.Nonce)
	return nil
}

type writer struct {
	conn   *conn
	closed atomic.Bool
	w      io.WriteCloser
	n      int64
}

func newWriter(c *conn, w io.WriteCloser) *writer {
	return &writer{
		conn: c,
		w:    w,
	}
}

func (w *writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.n += int64(n)
	return
}

func (w *writer) Close() error {
	if w.closed.CompareAndSwap(false, true) {
		w.w.Close()
		w.conn.Close()
	}
	return nil
}
