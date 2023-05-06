package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/infra"
)

const DriverName = "bt"

var _ infra.Driver = &Driver{}

type Driver struct {
	config Config
}

func (drv *Driver) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
