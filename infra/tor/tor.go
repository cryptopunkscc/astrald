package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/net/proxy"
	"log"
	"net"
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

	dialTimeout := &net.Dialer{Timeout: config.getDialTimeout()}
	tor.proxy, err = proxy.SOCKS5("tcp", config.getProxyAddress(), nil, dialTimeout)
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

func (tor Tor) Broadcast([]byte) error {
	return infra.ErrUnsupportedOperation
}

func (tor Tor) Scan(context.Context) (<-chan infra.Broadcast, <-chan error) {
	return nil, singleErrCh(infra.ErrUnsupportedOperation)
}

func (tor Tor) Announce(context.Context, id.Identity) error {
	return infra.ErrUnsupportedOperation
}

func (tor Tor) Discover(context.Context) (<-chan infra.Presence, error) {
	return nil, infra.ErrUnsupportedOperation
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
