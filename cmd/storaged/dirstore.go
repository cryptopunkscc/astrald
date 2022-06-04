package main

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"github.com/cryptopunkscc/astrald/proto/store"
	"os"
	"path/filepath"
)

var _ store.Store = &DirStore{}

type DirStore struct {
	dataDir string
}

func NewDirStore(dataDir string) *DirStore {
	return &DirStore{dataDir: dataDir}
}

func (s DirStore) Create(alloc uint64) (handler block.Block, tempID string, err error) {
	tempID = tempName(32)

	tempPath := filepath.Join(s.dataDir, tempID)

	file, err := os.Create(tempPath)
	if err != nil {
		return nil, "", err
	}

	return TempBlock{file, tempPath}, tempID, nil
}

func (s DirStore) Open(id data.ID, _ uint32) (block.Block, error) {
	fullpath := filepath.Join(s.dataDir, id.String())

	file, err := os.Open(fullpath)
	if err != nil {
		return nil, store.ErrNotFound
	}

	return block.Wrap(file), nil
}
