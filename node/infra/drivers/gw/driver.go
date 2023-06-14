package gw

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/infra"
)

const DriverName = "gw"
const ServiceName = "gateway"

var _ infra.Driver = &Driver{}

type Driver struct {
	config Config
	infra  infra.Infra
	log    *log.Logger
}

func (drv *Driver) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
