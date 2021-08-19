package adapter

import (
	"github.com/cryptopunkscc/astrald/components/serializer"
)

type reader struct {
	serializer.ReadCloser
	size int64
}

func (r reader) Size() (int64, error) {
	return r.size, nil
}