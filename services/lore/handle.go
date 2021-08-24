package lore

import (
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func (srv service) Observe(rc request.Context) (err error) {
	// Read type
	typ, err := rc.ReadStringWithSize8()
	if err != nil {
		log.Println(rc.Port, "cannot read type to observe", err)
		return
	}

	// Register observer
	srv.observers[rc.ReadWriteCloser] = typ
	log.Println(rc.Port, "added new observer for", typ)

	// Close blocking
	for {
		_, err = rc.ReadByte()
		if err != nil {
			log.Println(rc.Port, "removing file observer")
			delete(srv.observers, rc.ReadWriteCloser)
			return
		}
	}
}
