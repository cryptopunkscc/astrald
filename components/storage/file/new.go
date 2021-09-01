package file

import (
	"github.com/cryptopunkscc/astrald/components/storage"
)

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
