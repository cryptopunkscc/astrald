package main

// this file includes all network drivers that should be compiled into the node

import (
	_ "github.com/cryptopunkscc/astrald/node/infra/drivers/bt"
	_ "github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	_ "github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	_ "github.com/cryptopunkscc/astrald/node/infra/drivers/tor"
)
