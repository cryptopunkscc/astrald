package core

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ======================== File system resolver ========================

func newFsResolver(s storage) api.Resolver {
	return &fsResolver{s}
}

type fsResolver struct {
	storage
}

func (s *fsResolver) Reader(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (s *fsResolver) Info(path string) (files []api.Info, err error) {
	path = s.Absolute(path)
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

// ======================== External resolver api client ========================

func newRemoteResolver() api.Resolver {
	return &resolverClient{}
}

type resolverClient struct {
}

func (c *resolverClient) Reader(uri string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (c *resolverClient) Info(uri string) (files []api.Info, err error) {
	//TODO implement me
	panic("implement me")
}
