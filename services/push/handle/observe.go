package handle

import (
	"github.com/cryptopunkscc/astrald/services/util/handle"
)

func Observe(r *Request) (err error) {
	_ = handle.Observe(r.Context)
	return nil
}
