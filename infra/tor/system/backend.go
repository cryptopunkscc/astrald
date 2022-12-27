package system

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/infra/tor/tc"
	"github.com/cryptopunkscc/astrald/sig"
	"golang.org/x/net/proxy"
	"io"
	"net"
)

type Backend struct {
	config tor.Config
	proxy  proxy.ContextDialer
}

func (b *Backend) Run(ctx context.Context, config tor.Config) error {
	b.config = config

	var baseDialer = &net.Dialer{Timeout: config.GetDialTimeout()}

	socksProxy, err := proxy.SOCKS5("tcp", config.GetProxyAddress(), nil, baseDialer)
	if err != nil {
		return err
	}

	if dialContext, ok := socksProxy.(proxy.ContextDialer); !ok {
		return errors.New("type cast failed")
	} else {
		b.proxy = dialContext
	}

	<-ctx.Done()
	return nil
}

func (b *Backend) Dial(ctx context.Context, network string, addr string) (net.Conn, error) {
	//TODO: proxy is nil when Dial() is invoked before Run() sets up the proxy
	if b.proxy == nil {
		return nil, errors.New("proxy missing")
	}
	return b.proxy.DialContext(ctx, network, addr)
}

func (b *Backend) Listen(ctx context.Context, key tor.Key) (tor.Listener, error) {
	// Set up the listener for incoming tor connections
	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	sig.On(sig.Signal(ctx.Done()), func() {
		tcpListener.Close()
	})

	ctl, err := b.control()
	if err != nil {
		return nil, err
	}

	var keyStr string

	if len(key) == 0 {
		keyStr = tc.KeyNewV3
	} else {
		keyStr = "ED25519-V3:" + base64.StdEncoding.EncodeToString(key)
	}

	onion, err := ctl.AddOnion(keyStr, tc.Port(b.config.GetListenPort(), tcpListener.Addr().String()))
	if err != nil {
		return nil, err
	}

	return listener{tcpListener, onion, ctl}, nil
}

func (b *Backend) control() (*tc.Control, error) {
	conn, err := b.connect()
	if err != nil {
		return nil, err
	}

	ctl := tc.New(conn)
	if err := ctl.Authenticate(); err != nil {
		return nil, err
	}

	return ctl, nil
}

func (b *Backend) connect() (io.ReadWriteCloser, error) {
	return net.Dial("tcp", b.config.GetContolAddr())
}

func init() {
	tor.AddBackend("system", &Backend{})
}
