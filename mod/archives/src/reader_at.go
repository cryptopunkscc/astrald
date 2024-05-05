package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

const openTimeout = 15 * time.Second

type readerAt struct {
	objects  objects.Module
	objectID object.ID
	opts     *objects.OpenOpts
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	var opts = &objects.OpenOpts{}

	*opts = *r.opts
	opts.Offset = uint64(off)

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	f, err := r.objects.Open(ctx, r.objectID, opts)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Read(p)
}
