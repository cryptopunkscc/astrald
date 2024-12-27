package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) CreateObject(opts *objects.CreateOpts) (objects.Writer, error) {
	for _, dir := range mod.config.Store {
		r, err := mod.createObjectAt(dir, opts.Alloc)
		if err == nil {
			return r, err
		}
	}

	return nil, errors.New("no space available")
}

func (mod *Module) createObjectAt(path string, alloc int) (objects.Writer, error) {
	usage, err := DiskUsage(path)
	if err != nil {
		return nil, err
	}

	if usage.Free < uint64(alloc) {
		return nil, errors.New("not enough space left")
	}

	w, err := NewWriter(mod, path)

	return w, err
}
