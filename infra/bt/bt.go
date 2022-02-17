package bt

import (
	"github.com/cryptopunkscc/astrald/infra"
)

var _ infra.Network = &Bluetooth{}

type Bluetooth struct {
}

func New() (*Bluetooth, error) {
	return &Bluetooth{}, nil
}

func (Bluetooth) Name() string {
	return NetworkName
}
