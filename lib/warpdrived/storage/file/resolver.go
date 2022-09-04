package file

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

var _ storage.FileResolver = Resolver{}

type Resolver struct{}

func (s Resolver) Reader(path string, offset int64) (r io.ReadCloser, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	_, err = file.Seek(offset, 0)
	if err != nil {
		return
	}
	r = file
	return
}

func (s Resolver) Info(uri string) (files []warpdrive.Info, err error) {
	fn := func(uri string, info fs.FileInfo, err error) error {
		files = append(files, warpdrive.Info{
			Uri:   uri,
			Path:  uri,
			Size:  info.Size(),
			IsDir: info.IsDir(),
			Perm:  info.Mode().Perm(),
			Name:  path.Base(uri),
		})
		return nil
	}
	info, err := os.Lstat(uri)
	if err != nil {
		return
	}
	if info.IsDir() {
		err = filepath.Walk(uri, fn)
	} else {
		err = fn(uri, info, nil)
	}
	return
}
