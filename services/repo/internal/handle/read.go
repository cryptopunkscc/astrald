package handle

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/repo/internal/service"
	"io"
	"log"
)

func Read(c *service.Request) {

	// Read file id requested file fid
	log.Println(c.Port, "reading id")
	var idBuff [40]byte
	read, err := c.Read(idBuff[:])
	if err != nil || read != 40 {
		log.Println(c.Port, "cannot read id", read, err)
		return
	}
	id := fid.Unpack(idBuff)

	// Obtain file reader
	log.Println(c.Port, "getting reader for id", id.String())
	reader, err := c.Reader(id)
	if err != nil {
		log.Println(c.Port, "cannot get reader for id", id.String())
		return
	}
	// Send requested file
	log.Println(c.Port, "send file with id", id.String())
	defer func() { _ = c.Close() }()
	_, err = io.Copy(c, reader)

	log.Println(c.Port, "finalize read for", id.String(), err)
}
