package handle

import (
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func Observe(r *request.Context) (err error) {
	r.Observers[r] = struct{}{}
	log.Println(r.Port, "added new observer")
	for {
		if _, err = r.ReadByte(); err != nil {
			log.Println(r.Port, "removing observer")
			delete(r.Observers, r)
			return
		}
	}
}
