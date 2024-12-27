package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"os"
)

// OpenObject opens an object from the local filesystem
func (mod *Module) OpenObject(_ context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = defaultOpenOpts
	}

	if !opts.Zone.Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	paths := mod.path(objectID)
	for _, path := range paths {
		// check if the index for the path is valid
		if mod.validate(path) != nil {
			mod.enqueueUpdate(path)
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		n, err := f.Seek(int64(opts.Offset), io.SeekStart)
		if err != nil {
			f.Close()
			continue
		}

		if uint64(n) != opts.Offset {
			f.Close()
			continue
		}

		return &Reader{
			ReadSeekCloser: f,
			name:           path,
		}, nil
	}

	return nil, objects.ErrNotFound
}
