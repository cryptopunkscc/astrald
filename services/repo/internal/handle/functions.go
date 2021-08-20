package handle

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"io"
	"log"
)

func notifyObservers(c *service.Request, idPack [fid.Size]byte) {
	log.Println(c.Port, "notifying observers", len(c.Observers))
	for observer := range c.Observers {
		go func(w io.Writer) {
			_, err := w.Write(idPack[:])
			if err != nil {
				log.Println(c.Port, "cannot notify observer:", err)
			}
		}(observer)
	}
}
