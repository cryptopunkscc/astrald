package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
)

var _ storage.DataManager = &DataManager{}

var defaultReadOpts = &storage.ReadOpts{}

type DataManager struct {
	*Module

	readers sig.Map[string, storage.Reader]
	stores  sig.Map[string, storage.Store]
}

func NewDataManager(module *Module) *DataManager {
	return &DataManager{Module: module}
}

func (mod *DataManager) Read(id data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	if opts == nil {
		opts = defaultReadOpts
	}

	for _, reader := range mod.readers.Clone() {
		r, err := reader.Read(id, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (mod *DataManager) ReadAll(id data.ID, opts *storage.ReadOpts) ([]byte, error) {
	r, err := mod.Read(id, opts)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

func (mod *DataManager) Store(opts *storage.StoreOpts) (storage.DataWriter, error) {
	if opts == nil {
		opts = &storage.StoreOpts{}
	}

	for _, store := range mod.stores.Clone() {
		w, err := store.Store(opts)
		if err == nil {
			return NewDataWriterWrapper(mod.Module, w), err
		}
	}

	return nil, storage.ErrStorageUnavailable
}

func (mod *DataManager) StoreBytes(bytes []byte, opts *storage.StoreOpts) (data.ID, error) {
	if opts == nil {
		opts = &storage.StoreOpts{Alloc: len(bytes)}
	}

	w, err := mod.Store(opts)
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

func (mod *DataManager) AddReader(name string, reader storage.Reader) error {
	if mod.readers.Set(name, reader) {
		mod.events.Emit(storage.EventReaderAdded{
			Name:   name,
			Reader: reader,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}

func (mod *DataManager) RemoveReader(name string) error {
	if reader, ok := mod.readers.Delete(name); ok {
		mod.events.Emit(storage.EventReaderRemoved{
			Name:   name,
			Reader: reader,
		})
	}
	return storage.ErrNotFound
}

func (mod *DataManager) AddStore(name string, store storage.Store) error {
	if mod.stores.Set(name, store) {
		mod.events.Emit(storage.EventStoreAdded{
			Name:  name,
			Store: store,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}

func (mod *DataManager) RemoveStore(name string) error {
	if store, ok := mod.stores.Delete(name); ok {
		mod.events.Emit(storage.EventStoreRemoved{
			Name:  name,
			Store: store,
		})
	}
	return storage.ErrNotFound
}
