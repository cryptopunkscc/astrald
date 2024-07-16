package tor

import (
	"context"
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/mod/tor/tc"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	_net "net"
	"strings"
)

type Server struct {
	*Module
	endpoint *Endpoint
}

func NewServer(module *Module) *Server {
	return &Server{Module: module}
}

func (srv *Server) Run(ctx context.Context) error {
	key, err := srv.getPrivateKey()
	if err != nil {
		srv.log.Error("getPrivateKey: %s", err)
		return err
	}

	l, err := srv.listen(ctx, key)
	if err != nil {
		srv.log.Error("listen: %s", err)
		return err
	}
	defer l.Close()

	srv.endpoint, err = Parse(l.Addr())
	if err != nil {
		srv.log.Errorv(1, "error parsing tor key: %v", err)
	}

	srv.log.Log("listen %s", srv.endpoint)

	for {
		rawConn, err := l.Accept()
		switch {
		case err == nil:
		case strings.Contains(err.Error(), "use of closed network connection"):
			return err
		default:
			srv.log.Error("accept: %s", err)
			return err
		}

		var conn = newConn(rawConn, nil, false)

		go func() {
			_, err := srv.nodes.AcceptLink(ctx, conn)
			if err != nil {
				srv.log.Errorv(1, "handshake failed from %v: %v", conn.RemoteEndpoint(), err)
			}
		}()
	}

}

func (srv *Server) listen(ctx context.Context, key Key) (*listener, error) {
	// Set up the listener for incoming tor connections
	tcpListener, err := _net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	sig.On(ctx, func() {
		tcpListener.Close()
	})

	ctl, err := srv.control()
	if err != nil {
		return nil, err
	}

	var keyStr string

	if len(key) == 0 {
		keyStr = tc.KeyNewV3
	} else {
		keyStr = "ED25519-V3:" + base64.StdEncoding.EncodeToString(key)
	}

	onion, err := ctl.AddOnion(keyStr, tc.Port(srv.config.ListenPort, tcpListener.Addr().String()))
	if err != nil {
		return nil, err
	}

	return &listener{tcpListener, onion, ctl}, nil
}

func (srv *Server) control() (*tc.Control, error) {
	conn, err := srv.connect()
	if err != nil {
		return nil, err
	}

	ctl := tc.New(conn)
	if err := ctl.Authenticate(); err != nil {
		return nil, err
	}

	return ctl, nil
}

func (srv *Server) connect() (io.ReadWriteCloser, error) {
	return _net.Dial("tcp", srv.config.ControlAddr)
}

type listener struct {
	tcp   _net.Listener
	onion tc.Onion
	ctl   *tc.Control
}

func (l listener) Addr() string {
	return l.onion.ServiceID
}

func (l listener) PrivateKey() Key {
	s := l.onion.PrivateKey

	// force v3 as v2 is now considered insecure
	if !strings.HasPrefix(s, "ED25519-V3:") {
		return nil
	}

	ss := strings.Split(s, ":")

	if len(ss) != 2 {
		return nil
	}

	key, err := base64.StdEncoding.DecodeString(ss[1])
	if err != nil {
		return nil
	}

	return key
}

func (l listener) Accept() (_net.Conn, error) {
	return l.tcp.Accept()
}

func (l listener) Close() error {
	l.tcp.Close()

	return l.ctl.DelOnion(l.onion.ServiceID)
}
