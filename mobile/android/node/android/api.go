package android

import "io"

// Api interface provides access to native android API for golang modules.
type Api interface {
	Call(query string, arg interface{}) error
	Get(query string, arg interface{}, result interface{}) error
	Read(query string, arg interface{}) (io.ReadCloser, error)
}
