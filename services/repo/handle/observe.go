package handle

import (
	"github.com/cryptopunkscc/astrald/services/util/handle"
)

func Observe(c *Request) {
	_ = handle.Observe(&c.Context)
}
