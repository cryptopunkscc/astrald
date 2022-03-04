package storage

import (
	"io/ioutil"
	"path/filepath"
)

var _ Store = &FilesystemStorage{}

type FilesystemStorage struct {
	root string
}

func NewFilesystemStorage(path string) *FilesystemStorage {
	return &FilesystemStorage{root: path}
}

func (fs *FilesystemStorage) Root() string {
	return fs.root
}

func (fs *FilesystemStorage) LoadBytes(file string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(fs.root, file))
}

func (fs *FilesystemStorage) StoreBytes(file string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(fs.root, file), data, 0600)
}
