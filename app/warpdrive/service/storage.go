package warpdrive

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func receivedFiles() storage {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := home + "/warpdrive/incoming"
	err = os.Mkdir(dir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	return storage{dir}
}

func userFiles() storage {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return storage{dir}
}

type storage struct {
	root string
}

func (s storage) normalizePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(s.root, path)
}

func (s *storage) MkDir(path string, mode os.FileMode) error {
	return os.Mkdir(s.normalizePath(path), mode)
}

func (s *storage) Reader(path string) (io.ReadCloser, error) {
	return os.Open(s.normalizePath(path))
}

func (s *storage) Writer(path string, mode os.FileMode) (io.WriteCloser, error) {
	return os.OpenFile(s.normalizePath(path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
}

func (s *storage) Info(path string) (files []Info, err error) {
	fullPath := s.normalizePath(path)
	fn := func(path string, info fs.FileInfo, err error) error {
		path = strings.Replace(path, s.root, "", 1)
		path = strings.Replace(path, "/", "", 1)
		files = append(files, Info{
			Path:  path,
			Size:  info.Size(),
			IsDir: info.IsDir(),
			Perm:  info.Mode().Perm(),
		})
		return nil
	}
	info, err := os.Lstat(fullPath)
	if err != nil {
		return
	}
	if info.IsDir() {
		err = filepath.Walk(fullPath, fn)
	} else {
		err = fn(path, info, nil)
	}
	return
}
