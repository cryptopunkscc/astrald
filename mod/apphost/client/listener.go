package astral

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"net"
)

var _ net.Listener = &Listener{}

type Listener struct {
	listener net.Listener
	portName string
}

func NewListener(protocol string) (*Listener, error) {
	l, err := ipc.Listen(protocol)
	if err != nil {
		return nil, err
	}

	return &Listener{listener: l}, nil
}

func (l *Listener) NextQuery() (*Query, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	q := &Query{conn: conn}

	if err = cslq.Decode(conn, "v [c]c", &q.remoteID, &q.query); err != nil {
		conn.Close()
		return nil, err
	}

	return q, nil
}

func (l *Listener) QueryCh() <-chan *Query {
	ch := make(chan *Query, 1)

	go func() {
		defer close(ch)
		for {
			q, err := l.NextQuery()
			if err != nil {
				return
			}
			ch <- q
		}
	}()

	return ch
}

func (l *Listener) Accept() (net.Conn, error) {
	q, err := l.NextQuery()
	if err != nil {
		return nil, err
	}

	return q.Accept()
}

func (l *Listener) AcceptAll() <-chan net.Conn {
	ch := make(chan net.Conn, 0)

	go func() {
		defer close(ch)
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			ch <- conn
		}
	}()

	return ch
}

func (l *Listener) Close() error {
	return l.listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *Listener) Target() string {
	a := l.listener.Addr()
	return a.Network() + ":" + a.String()
}
