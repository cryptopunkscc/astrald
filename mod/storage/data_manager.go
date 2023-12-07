package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"slices"
	"time"
)

var _ storage.DataManager = &DataManager{}

type DataManager struct {
	*Module

	readers  sig.Set[storage.Reader]
	storers  sig.Set[storage.Storer]
	indexers sig.Set[storage.Indexer]
}

func NewDataManager(module *Module) *DataManager {
	return &DataManager{Module: module}
}

func (mod *DataManager) Read(id data.ID, opts *storage.ReadOpts) (io.ReadCloser, error) {
	for _, reader := range mod.readers.Clone() {
		r, err := reader.Read(id, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (mod *DataManager) Store(alloc int) (storage.DataWriter, error) {
	for _, storer := range mod.storers.Clone() {
		w, err := storer.Store(alloc)
		if err == nil {
			return w, err
		}
	}

	return nil, storage.ErrStorageUnavailable
}

func (mod *DataManager) IndexSince(since time.Time) []storage.DataInfo {
	var list []storage.DataInfo

	for _, indexer := range mod.indexers.Clone() {
		list = append(list, indexer.IndexSince(since)...)
	}

	slices.SortFunc(list, func(a, b storage.DataInfo) int {
		return a.IndexedAt.Compare(b.IndexedAt)
	})

	return list
}

func (mod *DataManager) AddReader(reader storage.Reader) {
	mod.readers.Add(reader)
}

func (mod *DataManager) RemoveReader(reader storage.Reader) {
	mod.readers.Remove(reader)
}

func (mod *DataManager) AddStorer(storer storage.Storer) {
	mod.storers.Add(storer)
}

func (mod *DataManager) RemoveStorer(storer storage.Storer) {
	mod.storers.Remove(storer)
}

func (mod *DataManager) AddIndexer(indexer storage.Indexer) {
	if err := mod.indexers.Add(indexer); err == nil {
		mod.events.Emit(storage.EventIndexerAdded{Indexer: indexer})
	}
}

func (mod *DataManager) RemoveIndexer(indexer storage.Indexer) {
	if err := mod.indexers.Remove(indexer); err == nil {
		mod.events.Emit(storage.EventIndexerRemoved{Indexer: indexer})
	}
}

func (mod *DataManager) Readers() []storage.Reader {
	return mod.readers.Clone()
}

func (mod *DataManager) Storers() []storage.Storer {
	return mod.storers.Clone()
}

func (mod *DataManager) Indexers() []storage.Indexer {
	return mod.indexers.Clone()
}
