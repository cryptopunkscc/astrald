package client

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"github.com/cryptopunkscc/astrald/components/sio"
)

type writer struct {
	sio.ReadWriteCloser
}

func (w writer) Finalize() (*fid.ID, error) {
	var idBuff [fid.Size]byte
	_, err := w.Read(idBuff[:])
	if err != nil {
		return nil, err
	}
	id := fid.Unpack(idBuff)
	return &id, nil
}
