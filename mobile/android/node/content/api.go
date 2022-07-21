package content

import (
	"io"
)

const (
	content = "sys/content"
	info    = "sys/content/info"
)

type Api interface {
	Reader(uri string) (io.ReadCloser, error)
	Info(uri string) (files Info, err error)
}

type Info struct {
	Uri  string
	Size int64
	Mime string
}
