package handle

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"io"
	"log"
)

func notifyObservers(c *Request, idPack [fid.Size]byte) {
	log.Println(c.Port, "notifying observers", len(c.Observers))
	for observer := range c.Observers {
		go func(w io.Writer) {
			_, err := w.Write(idPack[:])
			if err != nil {
				log.Println(c.Port, "cannot notify observer", err)
			}
		}(observer)
	}
}
