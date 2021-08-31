package file

import (
	"github.com/cryptopunkscc/astrald/components/storage"
	"github.com/cryptopunkscc/astrald/components/storage/compose"
)

func NewStorage(root string) storage.Storage {
	rw := NewReadWriteStorage(root)
	rm := NewReadMapStorage(root)
	r := compose.NewReadStorage(rw, rm)
	return compose.NewComposeStorage(r, rw, rm)
}

func NewReadStorage(root string) storage.ReadStorage {
	return fileStorage{dir: root}
}

func NewReadWriteDir(root string, dir string) storage.ReadWriteStorage {
	s, err := ResolveDir(root, dir)
	if err != nil {
		panic(err)
	}
	return fileStorage{dir: s}
}

func NewReadWriteStorage(root string) storage.ReadWriteStorage {
	dir, err := ResolveDir(root, "storage")
	if err != nil {
		panic(err)
	}
	return fileStorage{dir: dir}
}

func NewReadMapStorage(root string) storage.ReadMapStorage {
	storageDir, err := ResolveDir(root, "mappings")
	if err != nil {
		panic(err)
	}
	return &fileMapperStorage{dir: storageDir}
}
