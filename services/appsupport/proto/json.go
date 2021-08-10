package proto

import (
	"encoding/json"
	"io"
)

func NewJsonSocket(rwc io.ReadWriteCloser) *Socket {
	return &Socket{
		ReadWriteCloser: rwc,
		enc:             json.NewEncoder(rwc),
		dec:             json.NewDecoder(rwc),
	}
}
