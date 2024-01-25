package shares

import (
	"github.com/cryptopunkscc/astrald/mod/storage"
	"io"
)

var _ storage.DataReader = &RemoteDataReader{}

type RemoteDataReader struct {
	io.ReadCloser
}

func (r *RemoteDataReader) Info() *storage.ReaderInfo {
	return &storage.ReaderInfo{Name: "shares"}
}
