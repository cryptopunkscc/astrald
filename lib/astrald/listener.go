package astrald

import (
	"errors"
	"io"
	"net"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libapphost "github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

type Listener struct {
	net.Listener
	token  astral.Nonce
	doneCh chan struct{}
	done   atomic.Bool
}

var _ net.Listener = &Listener{}

func NewListener(protocol string, token astral.Nonce) (*Listener, error) {
	l, err := ipc.ListenAny(protocol)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Listener: l,
		doneCh:   make(chan struct{}),
		token:    token,
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	q, err := l.Next()
	if err != nil {
		return nil, err
	}

	return q.Accept(), nil
}

func (l *Listener) AcceptChannel(cfg ...channel.ConfigFunc) (*channel.Channel, error) {
	q, err := l.Next()
	if err != nil {
		return nil, err
	}

	return channel.New(q.Accept(), cfg...), nil
}

func (l *Listener) Next() (*PendingQuery, error) {
	// accent the next connection
	conn, err := l.Listener.Accept()
	if err != nil {
		l.setDone()
		return nil, err
	}
	ch := channel.New(conn)

	// read the request
	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *apphost.HandleQueryMsg:
		// check auth token
		if msg.AuthToken != l.token {
			ch.Send(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
			ch.Close()
			return nil, errors.New("invalid auth token")
		}

		return &PendingQuery{
			conn: conn,
			query: &astral.Query{
				Nonce:  msg.ID,
				Caller: msg.Caller,
				Target: msg.Target,
				Query:  string(msg.Query),
				Extra:  sig.Map[string, any]{},
			},
		}, nil

	case nil:
		return nil, err

	default:
		ch.Send(&apphost.ErrorMsg{Code: apphost.ErrCodeProtocolError})
		ch.Close()
		return nil, errors.New("unexpected message type " + msg.ObjectType())
	}
}

func (l *Listener) Close() error {
	l.setDone()
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}

func (l *Listener) String() string {
	a := l.Listener.Addr()
	return a.Network() + ":" + a.String()
}

func (l *Listener) Token() astral.Nonce {
	return l.token
}

func (l *Listener) SetToken(token astral.Nonce) {
	l.token = token
}

func (l *Listener) setDone() {
	if l.done.CompareAndSwap(false, true) {
		close(l.doneCh)
	}
}

func (l *Listener) Done() <-chan struct{} {
	return l.doneCh
}

func (l *Listener) Serve(ctx *astral.Context, router astral.Router) error {
	var errRejected *astral.ErrRejected

	for {
		q, err := l.Next()
		if err != nil {
			return err
		}

		var conn *libapphost.Conn

		w, err := router.RouteQuery(ctx, q.query, q.conn)
		switch {
		case err == nil:
			conn = q.Accept()

		case errors.As(err, &errRejected):
			q.RejectWithCode(int(errRejected.Code))
			continue

		default:
			q.Reject()
			continue
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()
	}
}
