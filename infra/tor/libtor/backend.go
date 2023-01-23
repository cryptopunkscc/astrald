package libtor

import (
	"context"
	bine "github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/ipsn/go-libtor"
	"net"
	"os"
)

type Backend struct {
	config tor.Config
	tor    *bine.Tor
}

func (b Backend) Dial(ctx context.Context, network string, addr string) (net.Conn, error) {
	dialer, err := b.tor.Dialer(ctx, nil)
	if err != nil {
		return nil, err
	}

	return dialer.Dial(network, addr)
}

func (b Backend) Listen(ctx context.Context, key tor.Key) (tor.Listener, error) {
	conf := &bine.ListenConf{
		Version3:    true,
		RemotePorts: []int{b.config.GetListenPort()},
	}

	if len(key) > 0 {
		conf.Key = ed25519.PrivateKey(key)
	}

	srv, err := b.tor.Listen(ctx, conf)

	if err != nil {
		return nil, err
	}

	return listener{srv}, nil
}

func (b *Backend) Run(ctx context.Context, config tor.Config) error {
	var err error

	b.config = config

	b.tor, err = bine.Start(
		ctx,
		&bine.StartConf{
			ProcessCreator:  libtor.Creator,
			DebugWriter:     nil,
			TempDataDirBase: os.TempDir(),
		},
	)

	if err != nil {
		log.Error("bine.Start: %s", err)
		return err
	}

	if err := b.tor.EnableNetwork(ctx, true); err != nil {
		return err
	}

	<-ctx.Done()

	return b.tor.Close()
}

func init() {
	tor.AddBackend("libtor", &Backend{})
}
