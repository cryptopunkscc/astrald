package content

import (
	"io"
	"os"
)

type Api interface {
	Reader(uri string) (io.ReadCloser, error)
	Info(uri string) (files []Info, err error)
}

type Info struct {
	Uri   string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}
