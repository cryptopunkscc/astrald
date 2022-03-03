package tor

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/storage"
	"golang.org/x/net/proxy"
)

const NetworkName = "tor"

var _ infra.Network = &Tor{}

type Tor struct {
	config Config
	store  storage.Store

	backend Backend

	proxy       proxy.Dialer
	serviceAddr Addr
}

func Run(ctx context.Context, config Config, store storage.Store) (*Tor, error) {
	var backendName = config.GetBackend()
	var backend, found = backends[backendName]

	if !found {
		return nil, errors.New("backend unavailable")
	}

	if err := backend.Run(ctx, config); err != nil {
		return nil, err
	}

	return &Tor{config: config, store: store, backend: backend}, nil
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
