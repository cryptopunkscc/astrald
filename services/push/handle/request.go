package handle

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/sync"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

type Request struct {
	request.Context
	Caller api.Identity
	Sync *sync.Client
}
