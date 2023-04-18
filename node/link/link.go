package link

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/cslq"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link/ctl"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync/atomic"
	"time"
)

const defaultIdleTimeout = 60 * time.Minute

const muxControlPort = 0

var ErrInvalidQuery = errors.New("invalid query")
var ErrRejected = errors.New("query rejected")

type QueryHandlerFunc func(query *Query) error

var log = _log.Tag("link")

// Link represents an astral link over an authenticated transport
type Link struct {
	activity      sig.Activity
	events        event.Queue
	conn          auth.Conn
	mux           *mux.FrameMux
	ctl           *ctl.Control
	conns         *ConnSet
	establishedAt time.Time
	ping          *Ping
	idle          *Idle
	control       *Control
	priority      int
	queryTimeout  time.Duration
	queryHandler  QueryHandlerFunc
	doneCh        chan struct{}
	closed        atomic.Bool
	err           error
}

func New(conn auth.Conn) *Link {
	l := &Link{
		conn:          conn,
		conns:         NewConnSet(),
		doneCh:        make(chan struct{}),
		establishedAt: time.Now().Round(0), // don't use monotonic clock
		queryTimeout:  DefaultQueryTimeout,
	}

	l.mux = mux.NewFrameMux(conn, l.onFrameDropped)

	// set up control connection
	var out = mux.NewFrameWriter(l.mux, muxControlPort)
	var in = mux.NewFrameReader()

	if err := l.mux.Bind(muxControlPort, in.HandleFrame); err != nil {
		panic(err)
	}

	l.ctl = ctl.New(streams.ReadWriteCloseSplit{Reader: in, Writer: out})

	l.control = NewControl(l)
	l.ping = NewPing(l)
	l.idle = NewIdle(l)

	return l
}

func (l *Link) Run(ctx context.Context) error {
	defer l.conn.Close() // close transport after link closes
	l.activity.Touch()   // reset idle to 0

	var runCtx, cancel = context.WithCancel(ctx)
	defer cancel()
	var group = tasks.Group(l.idle, l.ping, l.control, l.mux)
	group.DoneHandler = func(runner tasks.Runner, err error) {
		if runner == l.control {
			cancel()
		}
	}

	return group.Run(runCtx)
}

func (l *Link) Query(ctx context.Context, query string) (conn *Conn, err error) {
	if l.closed.Load() {
		return nil, ErrLinkClosed
	}

	if len(query) == 0 {
		return nil, ErrInvalidQuery
	}

	// silent queries do not affect activity
	if query[0] != '.' {
		l.activity.Add(1)
		defer l.activity.Done()
	}

	var reader = NewPortReader()
	localPort, err := l.mux.BindAny(reader)
	if err != nil {
		return nil, err
	}

	// write query frame
	err = l.ctl.WriteQuery(query, localPort)
	if err != nil {
		l.mux.Unbind(localPort)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// read the remote port of the connection
	var remotePort int
	var reply = make(chan struct{})

	go func() {
		err = cslq.Decode(reader, "s", &remotePort)
		close(reply)
	}()

	select {
	case <-ctx.Done():
		l.mux.Unbind(localPort)
		return nil, ctx.Err()

	case <-time.After(l.queryTimeout):
		l.mux.Unbind(localPort)
		return nil, ErrQueryTimeout

	case <-reply:
	}

	if err != nil {
		return nil, ErrRejected
	}

	conn = &Conn{
		localPort: localPort,
		query:     query,
		reader:    reader,
		writer:    mux.NewFrameWriter(l.mux, remotePort),
		link:      l,
		done:      make(chan struct{}),
		outbound:  true,
	}

	reader.SetErrorHandler(func(err error) {
		conn.closeWithError(err)
	})

	l.add(conn)

	l.Events().Emit(EventConnEstablished{Conn: conn})

	return
}

func (l *Link) Close() error {
	return l.CloseWithError(ErrLinkClosed)
}

func (l *Link) CloseWithError(e error) error {
	if l.closed.CompareAndSwap(false, true) {
		l.err = e
		_ = l.ctl.WriteClose()
		l.conn.Close()
		close(l.doneCh)
	}
	return nil
}

func (l *Link) onQuery(query string, remotePort int) error {
	var writer = mux.NewFrameWriter(l.mux, remotePort)
	var reader = NewPortReader()

	localPort, err := l.mux.BindAny(reader)
	if err != nil {
		writer.Close()
		return err
	}

	var q = &Query{
		query:     query,
		localPort: localPort,
		reader:    reader,
		writer:    writer,
		link:      l,
	}

	if l.queryHandler == nil {
		log.Log("no query handler set - automatically rejecting query: %s", query)
		_ = q.Reject()
		return nil
	}

	err = l.queryHandler(q)
	if err != nil {
		log.Log("rejecting query %s due to handler error: %s", query, err)
		q.Reject()
	}

	return err
}

func (l *Link) onDrop(remotePort int) error {
	conn := l.conns.FindByRemotePort(remotePort)
	if conn != nil {
		conn.Close()
	}
	return nil
}

func (l *Link) onClose() error {
	return l.CloseWithError(ErrLinkClosed)
}

func (l *Link) onFrameDropped(frame mux.Frame) error {
	if !frame.EOF() {
		_ = l.ctl.WriteDrop(frame.Port)
	}
	return nil
}

func (l *Link) onPing(port int) error {
	return l.mux.Close(port)
}

func (l *Link) add(conn *Conn) {
	if err := l.conns.Add(conn); err == nil {
		l.activity.Add(1)
	}
}

func (l *Link) remove(conn *Conn) error {
	if err := l.conns.Remove(conn); err == nil {
		l.activity.Done()
	}
	return nil
}
