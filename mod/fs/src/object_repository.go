package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Repository = &Repository{}

type Repository struct {
	mod  *Module
	name string
	path string
}

func (r Repository) Scan() (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	go func() {
		defer close(ch)

		var ids []*object.ID

		tx := r.mod.db.
			Model(&dbLocalFile{}).
			Select("data_id").
			Where("path like ?", r.path+"%").
			Find(&ids)

		if tx.Error != nil {
			return
		}

		for _, id := range ids {
			ch <- id
		}
	}()

	return ch, nil
}

func NewRepository(mod *Module, name string, path string) *Repository {
	return &Repository{mod: mod, name: name, path: path}
}

func (r Repository) Name() string {
	return r.name
}

func (r Repository) Create(opts *objects.CreateOpts) (objects.Writer, error) {
	if r.Free() < int64(opts.Alloc) {
		return nil, errors.New("not enough space left")
	}

	w, err := NewWriter(r.mod, r.path)

	return w, err
}

func (r Repository) Free() int64 {
	usage, err := DiskUsage(r.path)
	if err != nil {
		return -1
	}

	return int64(usage.Free)
}
