package handle

import (
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"io"
	"log"
)

func List(c *service.Request) {
	reader, err := c.List()
	if err != nil {
		log.Println(c.Port, "cannot list files", err)
		return
	}
	_, err = io.Copy(c, reader)
	if err != nil {
		log.Println(c.Port, "cannot send file ids")
		return
	}
}