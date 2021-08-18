package handle

import (
	"github.com/cryptopunkscc/astrald/services/util/request"
	"log"
)

func Observe(c *request.Context) {
	c.Observers[c] = struct{}{}
	log.Println(c.Port, "added new files observer")
	var buffer [1]byte
	for {
		_, err := c.Read(buffer[:])
		if err != nil {
			log.Println(c.Port, "removing file observer", err)
			delete(c.Observers, c)
			return
		}
	}
}
