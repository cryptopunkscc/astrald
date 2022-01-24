package local

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func NewStorage(path string) api.Storage {
	return storage{root: path}
}

type storage struct {
	root string
}

func (s storage) IsExist(err error) bool {
	return os.IsExist(err)
}

func (s storage) MkDir(path string, perm os.FileMode) error {
	return os.MkdirAll(s.normalizePath(path), perm)
}

func (s storage) FileWriter(path string, perm os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
}

func (s storage) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.root, path)
}
