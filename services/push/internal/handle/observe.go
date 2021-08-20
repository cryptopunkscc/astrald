package handle

import (
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"github.com/cryptopunkscc/astrald/services/util/handle"
)

func Observe(r *service.Request) (err error) {
	_ = handle.Observe(&r.Context.Context)
	return nil
}
