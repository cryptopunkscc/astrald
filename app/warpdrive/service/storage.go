package warpdrive

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func newStorage(path string) storage {
	return storage{root: path}
}

type storage struct {
	root string
}

func (s *storage) Absolute(relative string) string {
	return filepath.Join(s.root, relative)
}

func (s *storage) IsExist(err error) bool {
	return os.IsExist(err)
}

func (s *storage) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (s *storage) MkDir(path string, perm os.FileMode) error {
	return os.Mkdir(s.normalizePath(path), perm)
}

func (s *storage) Reader(path string) (io.ReadCloser, error) {
	return os.Open(s.normalizePath(path))
}

func (s *storage) Writer(path string, perm os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
}

func (s *storage) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.root, path)
}
