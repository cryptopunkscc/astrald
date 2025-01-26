package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"net"
	"sync/atomic"
)

var _ net.Listener = &Listener{}

type Listener struct {
	net.Listener
	token  string
	doneCh chan struct{}
	done   atomic.Bool
}

func NewListener(protocol string) (*Listener, error) {
	l, err := ipc.ListenAny(protocol)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Listener: l,
		doneCh:   make(chan struct{}),
	}, nil
}

func (l *Listener) Next() (*PendingQuery, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		l.setDone()
		return nil, err
	}

	var info apphost.QueryInfo

	_, err = info.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	if string(info.Token) != l.token {
		conn.Close()
		return nil, errors.New("token mismatch")
	}

	return &PendingQuery{
		conn: conn,
		info: &info,
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	q, err := l.Next()
	if err != nil {
		return nil, err
	}

	return q.Accept()
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

func (l *Listener) Token() string {
	return l.token
}

func (l *Listener) SetToken(token string) {
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
