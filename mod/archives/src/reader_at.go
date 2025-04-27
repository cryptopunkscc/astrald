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
	objectID object.ID
	opts     *objects.OpenOpts
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	var opts = &objects.OpenOpts{}

	*opts = *r.opts
	opts.Offset = uint64(off)

	ctx, cancel := astral.NewContext(nil).WithIdentity(r.identity).WithTimeout(openTimeout)
	defer cancel()

	f, err := r.objects.Open(ctx, r.objectID, opts)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Read(p)
}
