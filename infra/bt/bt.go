package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
)

const NetworkName = "bt"

var _ infra.Network = &Bluetooth{}

type Bluetooth struct {
}

func New() (*Bluetooth, error) {
	return &Bluetooth{}, nil
}

func (bt *Bluetooth) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (*Bluetooth) Name() string {
	return NetworkName
}
