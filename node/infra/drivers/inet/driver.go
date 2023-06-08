package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"sync"
)

var _ infra.Driver = &Driver{}

const DriverName = "inet"

type Driver struct {
	config      Config
	infra       infra.Infra
	log         *log.Logger
	publicAddrs []net.Endpoint
	mu          sync.Mutex
}

func (drv *Driver) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
