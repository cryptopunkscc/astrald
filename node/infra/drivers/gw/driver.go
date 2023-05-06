package gw

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/infra"
)

const DriverName = "gw"
const PortName = "gateway"

var _ infra.Driver = &Driver{}

type Driver struct {
	config Config
	infra  infra.Infra
}

func (drv *Driver) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
