package handle

import (
	"io"
	"log"
)

func List(c *Request) {
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
