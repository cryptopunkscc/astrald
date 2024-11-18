package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io/fs"
	"time"
)

const openTimeout = time.Second * 30

var _ fs.FS = &FS{}

type FS struct {
	mod objects.Module
	*objects.OpenOpts
}

func NewFS(mod objects.Module, openOpts *objects.OpenOpts) *FS {
	return &FS{mod: mod, OpenOpts: openOpts}
}

func (f *FS) Open(name string) (fs.File, error) {
	objectID, err := object.ParseID(name)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	r, err := f.mod.Open(ctx, objectID, f.OpenOpts)
	if err != nil {
		return nil, err
	}

	return &File{
		ID:     objectID,
		Reader: r,
	}, nil
}
