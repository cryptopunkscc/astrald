package handle

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/services/util/request"
	"io"
)

func Read(c *request.Context) {

	// Read file id requested file fid
	var idBuff [40]byte
	read, err := c.Read(idBuff[:])
	if err != nil || read != 40 {
		return
	}
	id := fid.Unpack(idBuff)

	// Obtain file reader
	reader, err := c.Reader(id)
	if err != nil {
		return
	}
	// Send requested file
	_, err = io.Copy(c, reader)
}
