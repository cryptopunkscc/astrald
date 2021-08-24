package handle

import (
	repo2 "github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

type Request struct {
	request.Context
	repo2.ReadWriteMapRepository
}

type Handle func(c *Request)

type Handlers map[byte]Handle
