package identifier

import (
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func (srv service) Observe(rc request.Context) (err error) {
	var query string

	// Read query
	if query, err = rc.ReadStringWithSize8(); err != nil {
		return
	}

	// Register observer
	srv.observers[rc.ReadWriteCloser] = query
	log.Println(Port, "added new observer for", query)

	// Close blocking
	for {
		if _, err = rc.ReadByte(); err != nil {
			log.Println(Port, "removing file observer")
			delete(srv.observers, rc.ReadWriteCloser)
			return
		}
	}
}
