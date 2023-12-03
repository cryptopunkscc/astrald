package storage

import (
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
)

func (mod *Module) AddReader(source storage.Reader) {
	mod.readersMu.Lock()
	defer mod.readersMu.Unlock()

	mod.readers = append(mod.readers, source)
}

func (mod *Module) RemoveReader(source storage.Reader) {
	mod.readersMu.Lock()
	defer mod.readersMu.Unlock()

	for i, s := range mod.readers {
		if s == source {
			mod.readers = append(mod.readers[:i], mod.readers[i+1:]...)
			return
		}
	}
}

func (mod *Module) Readers() []storage.Reader {
	mod.readersMu.Lock()
	defer mod.readersMu.Unlock()

	var list = make([]storage.Reader, 0, len(mod.readers))
	for _, source := range mod.readers {
		list = append(list, source)
	}
	return list
}
