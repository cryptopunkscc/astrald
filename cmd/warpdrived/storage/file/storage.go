package file

import (
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/core"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/storage"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ storage.File = Storage{}

type Storage core.Component

func (s Storage) IsExist(err error) bool {
	return os.IsExist(err)
}

func (s Storage) MkDir(path string, perm os.FileMode) error {
	return os.MkdirAll(s.normalizePath(path), perm)
}

func (s Storage) FileWriter(path string, perm os.FileMode) (io.WriteCloser, error) {
	// Try to create storage dir on demand.
	if err := s.MkDir("", 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
}

func (s Storage) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.StorageDir, path)
}
