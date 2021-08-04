package fs

import (
	"io/ioutil"
	"path/filepath"
)

type Filesystem struct {
	root string
}

func New(path string) *Filesystem {
	return &Filesystem{root: path}
}

func (fs *Filesystem) Read(file string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(fs.root, file))
}

func (fs *Filesystem) Write(file string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(fs.root, file), data, 0600)
}
