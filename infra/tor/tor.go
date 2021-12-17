package tor

import (
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/storage"
	"golang.org/x/net/proxy"
	"log"
	"net"
)

const NetworkName = "tor"

var _ infra.Network = &Tor{}

type Tor struct {
	config Config
	store  storage.Store

	proxy       proxy.Dialer
	serviceAddr Addr
}

func New(config Config, store storage.Store) *Tor {
	var err error
	var tor = &Tor{config: config, store: store}

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

func (tor Tor) Addresses() []infra.AddrSpec {
	if tor.serviceAddr.IsZero() {
		return nil
	}
	return []infra.AddrSpec{
		{
			Addr:   tor.serviceAddr,
			Global: true,
		},
	}
}
