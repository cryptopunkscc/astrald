package file

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ api.FileStorage = Storage{}

type Storage api.Core

func (s Storage) IsExist(err error) bool {
	return os.IsExist(err)
}

func (s Storage) MkDir(path string, perm os.FileMode) error {
	return os.MkdirAll(s.normalizePath(path), perm)
}

func (s Storage) FileWriter(path string, perm os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
}

func (s Storage) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.StorageDir, path)
}
