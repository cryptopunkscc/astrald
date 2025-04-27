package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const openTimeout = 15 * time.Second

type readerAt struct {
	identity *astral.Identity
	objects  objects.Module
	objectID *object.ID
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	ctx, cancel := astral.NewContext(nil).WithIdentity(r.identity).WithTimeout(openTimeout)
	defer cancel()

	f, err := r.objects.Root().Read(ctx, r.objectID, off, 0)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Read(p)
}
