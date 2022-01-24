package local

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func NewResolver() api.Resolver {
	return fsResolver{}
}

type fsResolver struct{}

func (s fsResolver) File(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (s fsResolver) Info(path string) (files []api.Info, err error) {
	fn := func(path string, info fs.FileInfo, err error) error {
		files = append(files, api.Info{
			Path:  path,
			Size:  info.Size(),
			IsDir: info.IsDir(),
			Perm:  info.Mode().Perm(),
		})
		return nil
	}
	info, err := os.Lstat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		err = filepath.Walk(path, fn)
	} else {
		err = fn(path, info, nil)
	}
	return
}
