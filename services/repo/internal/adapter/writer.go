package adapter

import (
	"github.com/cryptopunkscc/astrald/components/fid"
	"io"
)

type writer struct {
	io.ReadWriteCloser
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
