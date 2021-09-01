package auth

import "github.com/cryptopunkscc/astrald/api"

type Authorize func(core api.Core, conn api.ConnectionRequest) bool
