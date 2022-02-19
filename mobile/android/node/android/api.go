package android

import "io"

type Api interface {
	Call(query string, arg interface{}) error
	Get(query string, arg interface{}, result interface{}) error
	Read(query string, arg interface{}) (io.ReadCloser, error)
}
