package bt

import "github.com/cryptopunkscc/astrald/infra"

var Instance Client = &Bluetooth{}

type Client interface{ infra.Network }
