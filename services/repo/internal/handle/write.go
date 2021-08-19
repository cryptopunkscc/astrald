package handle

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"io"
	"log"
)

func Write(c *service.Request) {
	var sizeBuff [4]byte
	for {
		// Read next file size
		_, err := c.Read(sizeBuff[:])
		if err != nil {
			log.Println(c.Port, "closing stream for write request", err)
			return
		}
		log.Println(c.Port, "received bytes size:", sizeBuff)
		size := int64(binary.BigEndian.Uint32(sizeBuff[:]))
		log.Println(c.Port, "parsed size:", size)
		// Obtain file writer
		writer, err := c.Writer()
		if err != nil {
			log.Println(c.Port, "error while obtaining writer", err)
			return
		}
		log.Println(c.Port, "obtained writer")
		_, err = io.CopyN(writer, c, size)
		if err != nil {
			log.Println(c.Port, "cannot write to file")
			return
		}
		log.Println(c.Port, "successful write")
		id, err := writer.Finalize()
		if err != nil {
			log.Println(c.Port, "cannot finalize", err)
			return
		}
		idPack := id.Pack()
		notifyObservers(c, idPack)
		log.Println(c.Port, "sending")
		_, err = c.Write(idPack[:])
		if err != nil {
			log.Println(c.Port, "cannot write file fid", err)
			return
		}
	}
}
