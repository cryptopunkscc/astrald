package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/link/ctl"
	"github.com/cryptopunkscc/astrald/query"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync/atomic"
	"time"
)

const defaultIdleTimeout = 60 * time.Minute
const muxControlPort = 0

var _ query.Link = &Link{}

// Link represents an astral link over an authenticated transport
type Link struct {
	activity      sig.Activity
	events        events.Queue
	transport     net.SecureConn
	mux           *mux.FrameMux
	ctl           *ctl.Control
	conns         *ConnSet
	establishedAt time.Time
	health        *Health
	idle          *Idle
	control       *Control
	priority      int
	queryTimeout  time.Duration
	queryRouter   query.Router
	doneCh        chan struct{}
	closed        atomic.Bool
	err           error
	log           *log.Logger
}

func New(transport net.SecureConn, l *log.Logger) *Link {
	lnk := &Link{
		transport:     transport,
		conns:         NewConnSet(),
		doneCh:        make(chan struct{}),
		establishedAt: time.Now().Round(0), // don't use monotonic clock
		queryTimeout:  DefaultQueryTimeout,
		log:           l,
	}

	lnk.mux = mux.NewFrameMux(transport, lnk.onFrameDropped)

	// set up control connection
	var out = mux.NewFrameWriter(lnk.mux, muxControlPort)
	var in = mux.NewFrameReader()

	if err := lnk.mux.Bind(muxControlPort, in.HandleFrame); err != nil {
		panic(err)
	}

	lnk.ctl = ctl.New(streams.ReadWriteCloseSplit{Reader: in, Writer: out})
	lnk.control = NewControl(lnk)
	lnk.health = NewHealth(lnk)
	lnk.idle = NewIdle(lnk)

	return lnk
}

func (l *Link) Run(ctx context.Context) error {
	defer l.transport.Close() // close transport after link closes
	l.activity.Touch()        // reset idle to 0

	var runCtx, cancel = context.WithCancel(ctx)
	defer cancel()
	var group = tasks.Group(l.idle, l.health, l.control, l.mux)
	group.DoneHandler = func(runner tasks.Runner, err error) {
		if runner == l.control {
			l.CloseWithError(err)
			cancel()
		}
	}

	return group.Run(runCtx)
}

func (l *Link) handleQueryMessage(ctx context.Context, msg ctl.QueryMessage) (err error) {
	if l.queryRouter == nil {
		return errors.New("query handler missing")
	}

	var q = query.NewOrigin(l.RemoteIdentity(), l.LocalIdentity(), msg.Query(), query.OriginNetwork)

	var localWriter = &WriterMonitor{Target: NewSecureFrameWriter(l, msg.Port())}

	remoteWriter, err := l.queryRouter.RouteQuery(ctx, q, localWriter)
	if err != nil {
		localWriter.Close()
		return nil
	}

	remoteWriter = &WriterMonitor{Target: remoteWriter}

	localPort, err := l.mux.BindAny(WriterFrameHandler{remoteWriter})
	if err != nil {
		localWriter.Close()
		return err
	}

	conn := NewConn(localPort, localWriter, msg.Port(), remoteWriter, msg.Query(), false)

	l.add(conn)

	return cslq.Encode(localWriter, "s", localPort)
}

func (l *Link) Close() error {
	return l.CloseWithError(ErrLinkClosed)
}

func (l *Link) CloseWithError(e error) error {
	if l.closed.CompareAndSwap(false, true) {
		l.err = e
		_ = l.ctl.WriteClose()
		l.transport.Close()
		close(l.doneCh)
	} else {
		l.log.Errorv(2, "link with %s over %s double closed. first: %s, new: %s",
			l.RemoteIdentity(),
			l.Network(),
			l.err,
			e,
		)
	}
	return nil
}

func (l *Link) Transport() net.SecureConn {
	return l.transport
}

func (l *Link) onDrop(remotePort int) error {
	conn := l.conns.FindByRemotePort(remotePort)
	if conn != nil {
		conn.localWriter.Close()
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

		l.events.Emit(EventConnAdded{
			localName:  l.log.Sprintf("%v", conn.LocalIdentity()),
			remoteName: l.log.Sprintf("%v", conn.RemoteIdentity()),
			Conn:       conn,
		})

		conn.StateChanged = func() {
			if conn.State() == StateClosed {
				l.remove(conn)

				l.events.Emit(EventConnRemoved{
					localName:  l.log.Sprintf("%v", conn.LocalIdentity()),
					remoteName: l.log.Sprintf("%v", conn.RemoteIdentity()),
					Conn:       conn,
				})
			}
		}
	}
}

func (l *Link) remove(conn *Conn) error {
	if err := l.conns.Remove(conn); err == nil {
		l.activity.Done()
	}
	return nil
}
