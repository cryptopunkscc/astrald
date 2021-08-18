package handle

import (
	"github.com/cryptopunkscc/astrald/services/push/internal/service"
	"log"
)

func Observe(r *service.Request) (err error) {
	r.Observers[r] = struct{}{}
	for {
		if _, err = r.ReadByte(); err != nil {
			log.Println(r.Port, "removing file observer")
			delete(r.Observers, r)
			return
		}
	}
}
