package tor

import (
	"context"
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/tor/tc"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	_net "net"
	"strings"
)

var _ infra.Listener = &Driver{}

func (drv *Driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	output := make(chan net.Conn)

	// process incoming connections
	go func() {
		defer close(output)

		key, err := drv.getPrivateKey()
		if err != nil {
			log.Error("getPrivateKey: %s", err)
			return
		}

		l, err := drv.listen(ctx, key)
		if err != nil {
			log.Error("listen: %s", err)
			return
		}
		defer l.Close()

		drv.serviceAddr, _ = Parse(l.Addr())

		log.Log("listen %s", drv.serviceAddr)

		for {
			conn, err := l.Accept()
			switch {
			case err == nil:
			case strings.Contains(err.Error(), "use of closed network connection"):
				return
			default:
				log.Error("accept: %s", err)
				return
			}
			output <- newConn(conn, Endpoint{}, false)
		}
	}()

	return output, nil
}

func (drv *Driver) listen(ctx context.Context, key Key) (*listener, error) {
	// Set up the listener for incoming tor connections
	tcpListener, err := _net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	sig.On(ctx, func() {
		tcpListener.Close()
	})

	ctl, err := drv.control()
	if err != nil {
		return nil, err
	}

	var keyStr string

	if len(key) == 0 {
		keyStr = tc.KeyNewV3
	} else {
		keyStr = "ED25519-V3:" + base64.StdEncoding.EncodeToString(key)
	}

	onion, err := ctl.AddOnion(keyStr, tc.Port(drv.config.ListenPort, tcpListener.Addr().String()))
	if err != nil {
		return nil, err
	}

	return &listener{tcpListener, onion, ctl}, nil
}

func (drv *Driver) control() (*tc.Control, error) {
	conn, err := drv.connect()
	if err != nil {
		return nil, err
	}

	ctl := tc.New(conn)
	if err := ctl.Authenticate(); err != nil {
		return nil, err
	}

	return ctl, nil
}

func (drv *Driver) connect() (io.ReadWriteCloser, error) {
	return _net.Dial("tcp", drv.config.ControlAddr)
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
