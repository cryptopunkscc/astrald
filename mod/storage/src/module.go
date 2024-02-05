package storage

import (
	"cmp"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
	"slices"
)

var _ storage.Module = &Module{}
var defaultReadOpts = &storage.OpenOpts{Virtual: true}

// ReadAllMaxSize is the limit on data size accepted by ReadAll() (to avoid accidental OOM)
var ReadAllMaxSize uint64 = 1024 * 1024 * 1024

type Opener struct {
	storage.Opener
	Name     string
	Priority int
}

type Creator struct {
	storage.Creator
	Name     string
	Priority int
}

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context

	openers  sig.Map[string, *Opener]
	creators sig.Map[string, *Creator]
}

func (mod *Module) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	if opts == nil {
		opts = defaultReadOpts
	}

	openers := mod.openers.Values()

	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	for _, opener := range openers {
		r, err := opener.Open(dataID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	if opts == nil {
		opts = &storage.CreateOpts{}
	}

	if opts.Alloc < 0 {
		return nil, errors.New("alloc cannot be less than 0")
	}

	creators := mod.creators.Values()

	slices.SortFunc(creators, func(a, b *Creator) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	for _, creator := range creators {
		w, err := creator.Create(opts)
		if err == nil {
			return NewDataWriterWrapper(mod, w), err
		}
	}

	return nil, storage.ErrStorageUnavailable
}

func (mod *Module) ReadAll(id data.ID, opts *storage.OpenOpts) ([]byte, error) {
	if id.Size > ReadAllMaxSize {
		return nil, errors.New("data too big")
	}
	r, err := mod.Open(id, opts)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

func (mod *Module) Put(bytes []byte, opts *storage.CreateOpts) (data.ID, error) {
	if opts == nil {
		opts = &storage.CreateOpts{Alloc: len(bytes)}
	}

	w, err := mod.Create(opts)
	if err != nil {
		return data.ID{}, err
	}
	defer w.Discard()

	_, err = w.Write(bytes)
	if err != nil {
		return data.ID{}, err
	}

	return w.Commit()
}

func (mod *Module) AddOpener(name string, opener storage.Opener, priority int) error {
	_, ok := mod.openers.Set(name, &Opener{
		Opener:   opener,
		Name:     name,
		Priority: priority,
	})
	if ok {
		mod.events.Emit(storage.EventOpenerAdded{
			Name:   name,
			Opener: opener,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}

func (mod *Module) AddCreator(name string, creator storage.Creator, priority int) error {
	_, ok := mod.creators.Set(name, &Creator{
		Creator:  creator,
		Name:     name,
		Priority: priority,
	})

	if ok {
		mod.events.Emit(storage.EventStoreAdded{
			Name:    name,
			Creator: creator,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}

func (mod *Module) RemoveOpener(name string) error {
	if opener, ok := mod.openers.Delete(name); ok {
		mod.events.Emit(storage.EventReaderRemoved{
			Name:   name,
			Opener: opener,
		})
	}
	return storage.ErrNotFound
}

func (mod *Module) RemoveCreator(name string) error {
	if creator, ok := mod.creators.Delete(name); ok {
		mod.events.Emit(storage.EventStoreRemoved{
			Name:    name,
			Creator: creator,
		})
	}
	return storage.ErrNotFound
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	<-ctx.Done()

	return nil
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}
