package tor

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/net/proxy"
)

const NetworkName = "tor"

var _ infra.Network = &Tor{}

type Tor struct {
	config      Config
	rootDir     string
	backend     Backend
	proxy       proxy.Dialer
	serviceAddr Addr
}

func New(config Config, rootDir string) (*Tor, error) {
	tor := &Tor{
		config:  config,
		rootDir: rootDir,
	}

	var backendName = tor.config.GetBackend()
	var backend, found = backends[backendName]

	if !found {
		return nil, errors.New("backend unavailable")
	}

	tor.backend = backend

	return tor, nil
}

func (tor *Tor) Run(ctx context.Context) error {
	return tor.backend.Run(ctx, tor.config)
}

// Name returns the network name
func (tor *Tor) Name() string {
	return NetworkName
}

func (tor *Tor) Addresses() []infra.AddrSpec {
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
