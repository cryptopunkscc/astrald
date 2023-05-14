package astral

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"net"
)

var _ net.Listener = &Listener{}

type Listener struct {
	listener net.Listener
	portName string
	onClose  func()
}

func newListener(protocol string) (*Listener, error) {
	l, err := proto.ListenAny(protocol)
	if err != nil {
		return nil, err
	}

	return &Listener{listener: l}, nil
}

func (l *Listener) Next() (*QueryData, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	var in proto.InQueryParams

	if err = cslq.Decode(conn, "v", &in); err != nil {
		conn.Close()
		return nil, err
	}

	q := &QueryData{
		conn:     conn,
		query:    in.Query,
		remoteID: in.Identity,
	}

	return q, nil
}

func (l *Listener) QueryCh() <-chan *QueryData {
	ch := make(chan *QueryData, 1)

	go func() {
		defer close(ch)
		for {
			q, err := l.Next()
			if err != nil {
				return
			}
			ch <- q
		}
	}()

	return ch
}

func (l *Listener) Accept() (net.Conn, error) {
	q, err := l.Next()
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
	if l.onClose != nil {
		l.onClose()
	}
	return l.listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *Listener) Target() string {
	a := l.listener.Addr()
	return a.Network() + ":" + a.String()
}
