package proto

import (
	"encoding/gob"
	"io"
)

func NewGobSocket(rwc io.ReadWriteCloser) *Socket {
	return &Socket{
		ReadWriteCloser: rwc,
		enc:             gob.NewEncoder(rwc),
		dec:             gob.NewDecoder(rwc),
	}
}
