package adapter

import (
	"github.com/cryptopunkscc/astrald/components/sio"
)

type reader struct {
	sio.ReadCloser
	size int64
}

func (r reader) Size() (int64, error) {
	return r.size, nil
}