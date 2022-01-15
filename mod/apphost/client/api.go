package astral

import "io"

type Api interface {
	Register(name string) (Port, error)
	Query(nodeID string, query string) (io.ReadWriteCloser, error)
}

type Port interface {
	Next() <-chan Request
	Close() error
}
type Request interface {
	Caller() string
	Query() string
	Accept() (io.ReadWriteCloser, error)
	Reject()
}
