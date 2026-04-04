package fs

import (
	"io/fs"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

const openTimeout = time.Second * 15

type FS struct {
	repo     objects.Repository
	identity *astral.Identity
}

var _ fs.FS = &FS{}

func NewFS(repo objects.Repository) *FS {
	return &FS{
		repo: repo,
	}
}

func (f *FS) Open(name string) (fs.File, error) {
	return f.OpenContext(astral.NewContext(nil).WithZone(astral.ZoneAll), name)
}

func (f *FS) OpenContext(ctx *astral.Context, name string) (fs.File, error) {
	// parse object id
	objectID, err := astral.ParseID(name)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	// set a timeout for the operation
	ctx, cancel := ctx.WithTimeout(openTimeout)
	defer cancel()

	// open the file
	r, err := f.repo.Read(ctx, objectID, 0, 0)
	if err != nil {
		return nil, err
	}

	// wrap the reader into a
	return &File{
		ID:         objectID,
		ReadCloser: objects.NewReadSeeker(ctx, objectID, f.repo, r),
	}, nil
}
