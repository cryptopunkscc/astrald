package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io/fs"
	"time"
)

const openTimeout = time.Second * 30

type FS struct {
	ctx      *astral.Context
	identity *astral.Identity
	mod      objects.Module
}

var _ fs.FS = &FS{}

func NewFS(ctx *astral.Context, mod objects.Module) *FS {
	return &FS{
		ctx: ctx,
		mod: mod,
	}
}

func (f *FS) Open(name string) (fs.File, error) {
	objectID, err := object.ParseID(name)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	ctx, cancel := f.ctx.WithTimeout(openTimeout)
	defer cancel()

	r, err := f.mod.Root().Read(ctx, objectID, 0, 0)
	if err != nil {
		return nil, err
	}

	rs := objects.NewReadSeeker(f.ctx, objectID, f.mod.Root(), r)

	return &File{
		ID:     objectID,
		Reader: rs,
	}, nil
}
