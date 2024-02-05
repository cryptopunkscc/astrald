package zip

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

type readerAt struct {
	storage storage.Module
	dataID  data.ID
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	f, err := r.storage.Open(r.dataID, &storage.OpenOpts{Offset: uint64(off), Virtual: true})
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Read(p)
}
