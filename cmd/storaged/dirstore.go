package main

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"github.com/cryptopunkscc/astrald/proto/store"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
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

func (s DirStore) Download(blockID data.ID, offset uint64, limit uint64) (io.ReadCloser, error) {
	fullpath := filepath.Join(s.dataDir, blockID.String())

	file, err := os.Open(fullpath)
	if err != nil {
		return nil, store.ErrNotFound
	}

	if offset != 0 {
		if _, err := file.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
	}

	if limit > 0 {
		return &streams.LimitedReader{
			ReadCloser: file,
			Limit:      limit,
		}, nil
	}

	return file, nil
}
