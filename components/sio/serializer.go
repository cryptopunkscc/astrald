package sio

import "io"

type readWriteCloser struct {
	io.Closer
	Reader
	Writer
}
