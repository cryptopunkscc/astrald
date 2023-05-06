package tor

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/infra"
	"golang.org/x/net/proxy"
	"net"
)

const DriverName = "tor"

var _ infra.Driver = &Driver{}

type Driver struct {
	config      Config
	configStore config.Store
	proxy       proxy.ContextDialer
	serviceAddr Endpoint
}

func (drv *Driver) Run(ctx context.Context) error {
	var baseDialer = &net.Dialer{Timeout: drv.config.DialTimeout}

	socksProxy, err := proxy.SOCKS5("tcp", drv.config.TorProxy, nil, baseDialer)
	if err != nil {
		return err
	}

	if dialContext, ok := socksProxy.(proxy.ContextDialer); !ok {
		return errors.New("type cast failed")
	} else {
		drv.proxy = dialContext
	}

	<-ctx.Done()
	return nil
}
