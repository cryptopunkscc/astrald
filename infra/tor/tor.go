package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"io/ioutil"
	"log"
	"net"
)

const NetworkName = "tor"
const torPort = 1791

var _ infra.Network = &Tor{}

type Tor struct {
	serviceAddr Addr
}

func (tor Tor) Name() string {
	return NetworkName
}

func (tor Tor) Unpack(bytes []byte) (infra.Addr, error) {
	return Unpack(bytes)
}

func (tor *Tor) Listen(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	output, errCh := make(chan infra.Conn), make(chan error)

	ctl, err := Open()
	if err != nil {
		return nil, errChan(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, errChan(err)
	}

	key := ""
	newKey := true

	bytes, err := ioutil.ReadFile(keyPath())
	if err == nil {
		key = string(bytes)
		newKey = false
	}

	srv, err := ctl.StartService(key, map[int]string{torPort: listener.Addr().String()})
	if err != nil {
		return nil, errChan(err)
	}

	tor.serviceAddr, _ = Parse(srv.serviceID)

	log.Printf("listen tor %s.onion:%d\n", srv.serviceID, torPort)
	if newKey {
		ioutil.WriteFile(keyPath(), []byte(srv.PrivateKey()), 0600)
	}

	go func() {
		defer close(output)

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			output <- newConn(conn, Addr{}, false)
		}

		log.Printf("closed tor %s.onion:%d\n", srv.serviceID, torPort)

		srv.Close()
	}()

	return output, errCh
}

func (tor Tor) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	a, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	return Dial(ctx, a)
}

func (tor Tor) Advertise(ctx context.Context, payload []byte) <-chan error {
	return errChan(infra.ErrUnsupportedOperation)
}

func (tor Tor) Scan(ctx context.Context) (<-chan infra.Ad, <-chan error) {
	return nil, errChan(infra.ErrUnsupportedOperation)
}

func (tor Tor) Addresses() []infra.AddrDesc {
	if tor.serviceAddr.IsZero() {
		return nil
	}
	return []infra.AddrDesc{
		{
			Addr:   tor.serviceAddr,
			Public: true,
		},
	}
}

func New() infra.Network {
	return &Tor{}
}

func errChan(err error) <-chan error {
	ch := make(chan error, 1)
	defer close(ch)
	ch <- err
	return ch
}
