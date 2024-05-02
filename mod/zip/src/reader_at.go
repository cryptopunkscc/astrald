package zip

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

type readerAt struct {
	objects  objects.Module
	objectID object.ID
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	f, err := r.objects.Open(r.objectID, &objects.OpenOpts{Offset: uint64(off), Virtual: true})
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Read(p)
}
