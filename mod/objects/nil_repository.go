package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

type NilRepository struct {
}

func (n NilRepository) Create(ctx *astral.Context, opts *CreateOpts) (Writer, error) {
	return nil, errors.ErrUnsupported
}

func (n NilRepository) Contains(ctx *astral.Context, objectID *object.ID) (bool, error) {
	return false, errors.ErrUnsupported
}

func (n NilRepository) Scan(ctx *astral.Context, follow bool) (<-chan *object.ID, error) {
	return nil, errors.ErrUnsupported
}

func (n NilRepository) Delete(ctx *astral.Context, objectID *object.ID) error {
	return errors.ErrUnsupported
}

func (n NilRepository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (Reader, error) {
	return nil, errors.ErrUnsupported
}

func (n NilRepository) Free(ctx *astral.Context) (int64, error) {
	return 0, errors.ErrUnsupported
}
