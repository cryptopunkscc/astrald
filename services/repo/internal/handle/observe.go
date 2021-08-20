package handle

import (
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"github.com/cryptopunkscc/astrald/services/util/handle"
)

func Observe(c *service.Request) {
	_ = handle.Observe(&c.Context.Context)
}
