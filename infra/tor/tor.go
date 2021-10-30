package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/net/proxy"
	"log"
)

const NetworkName = "tor"

var _ infra.Network = &Tor{}

type Tor struct {
	config Config

	proxy       proxy.Dialer
	serviceAddr Addr
}

func New(config Config) *Tor {
	var err error
	var tor = &Tor{config: config}

	tor.proxy, err = proxy.SOCKS5("tcp", config.getProxyAddress(), nil, nil)
	if err != nil {
		log.Println("tor: config error:", err)
		return nil
	}

	return tor
}

// Name returns the network name
func (tor Tor) Name() string {
	return NetworkName
}

// Unpack deserializes an address from its binary format
func (tor Tor) Unpack(bytes []byte) (infra.Addr, error) {
	return Unpack(bytes)
}

func (tor Tor) Broadcast(ctx context.Context, payload []byte) <-chan error {
	return singleErrCh(infra.ErrUnsupportedOperation)
}

func (tor Tor) Scan(ctx context.Context) (<-chan infra.Broadcast, <-chan error) {
	return nil, singleErrCh(infra.ErrUnsupportedOperation)
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

// singleErrCh creates a closed error channel with a single error in it
func singleErrCh(err error) <-chan error {
	ch := make(chan error, 1)
	defer close(ch)
	ch <- err
	return ch
}
