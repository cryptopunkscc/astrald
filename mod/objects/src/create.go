package objects

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

const MaxAlloc int64 = 1 * 1024 * 1024 * 1024 * 1024 //1TB; gomobile requires explicit int64 type.

type Creator struct {
	objects.Repository
	Priority int
}

func (mod *Module) Create(o *objects.CreateOpts) (objects.Writer, error) {
	var opts objects.CreateOpts
	if o != nil {
		opts = *o
	}

	if opts.Alloc < 0 {
		return nil, errors.New("alloc cannot be less than 0")
	}

	if int64(opts.Alloc) > MaxAlloc {
		return nil, errors.New("alloc exceeds limit")
	}

	if opts.As.IsZero() {
		opts.As = mod.node.Identity()
	}

	if opts.Repo == "" {
		opts.Repo = "default"
	}

	repo, ok := mod.repos.Get(opts.Repo)
	if !ok {
		return nil, fmt.Errorf("repo %s not found", opts.Repo)
	}

	w, err := repo.Create(&opts)
	if err != nil {
		return nil, err
	}

	return NewWriterWrapper(mod, w), err
}

func (mod *Module) AddRepository(repo objects.Repository) error {
	_, ok := mod.repos.Set(repo.Name(), repo)
	if !ok {
		return fmt.Errorf("repo %s already added", repo.Name())
	}

	return nil
}
