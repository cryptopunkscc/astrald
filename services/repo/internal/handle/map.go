package handle

import (
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"log"
)

func Map(c *service.Request) {
	for {
		log.Println(c.Port, "reading path to map")
		path, err := c.ReadStringWithSize16()
		if err != nil {
			log.Println(c.Port, "cannot read path", err)
			return
		}

		go func() {
			log.Println(c.Port, "mapping path", path)
			id, err := c.Map(path)
			if err != nil {
				log.Println(c.Port, "cannot map path", path, err)
				c.Close()
				return
			}
			idPack := id.Pack()
			notifyObservers(c, id.Pack())
			_, err = c.Write(idPack[:])
			if err != nil {
				log.Println(c.Port, "cannot notify caller")
				return
			}
		}()
	}
}
