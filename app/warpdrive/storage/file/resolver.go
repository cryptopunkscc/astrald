package file

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var _ api.FileResolver = Resolver{}

type Resolver struct{}

func (s Resolver) Reader(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (s Resolver) Info(path string) (files []api.Info, err error) {
	fn := func(path string, info fs.FileInfo, err error) error {
		files = append(files, api.Info{
			Uri:   path,
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
