package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
	"net"
)

type AuthFunc func(identity id.Identity, query string) bool

type Listener struct {
	listener   net.Listener
	portCloser io.Closer
	authFunc   AuthFunc
	portName   string
}

func NewListener(protocol string) (*Listener, error) {
	l, err := ipc.Listen(protocol)
	if err != nil {
		return nil, err
	}

	return &Listener{listener: l}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return nil, err
		}

		stream := cslq.NewEndec(conn)

		var (
			remoteID id.Identity
			query    string
		)

		if err := stream.Decode("v [c]c", &remoteID, &query); err != nil {
			conn.Close()
			return nil, err
		}

		if (l.authFunc == nil) || l.authFunc(remoteID, query) {
			stream.Encode("c", proto.ResponseOK)

			return Conn{Conn: conn, remoteAddr: Addr{remoteID.String()}}, nil
		}

		stream.Encode("c", proto.ResponseRejected)
		conn.Close()
	}
}

func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *Listener) Auth(authFunc AuthFunc) *Listener {
	return &Listener{
		listener:   l.listener,
		portCloser: l.portCloser,
		authFunc:   authFunc,
	}
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
	l.listener.Close()
	return l.portCloser.Close()
}

func (l *Listener) Target() string {
	a := l.listener.Addr()
	return a.Network() + ":" + a.String()
}
