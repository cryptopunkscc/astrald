package bt

import (
	"github.com/cryptopunkscc/astrald/infra"
)

const NetworkName = "bt"

var _ infra.Network = &Bluetooth{}

type Bluetooth struct {
}

func New() (*Bluetooth, error) {
	return &Bluetooth{}, nil
}

func (Bluetooth) Name() string {
	return NetworkName
}
