package handle

import (
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"log"
)

func Observe(c *service.Request) {
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
