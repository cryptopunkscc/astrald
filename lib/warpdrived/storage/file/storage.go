package file

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage"
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

func (s Storage) FileWriter(path string, perm os.FileMode, offset int64) (w io.WriteCloser, err error) {
	// Try to create storage dir on demand.
	if err = s.MkDir("", 0755); err != nil {
		return
	}
	file, err := os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return
	}
	_, err = file.Seek(offset, 0)
	if err != nil {
		return
	}
	w = file
	return
}

func (s Storage) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.StorageDir, path)
}
