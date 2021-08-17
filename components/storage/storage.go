package storage

import "io"

type Storage interface {
	Reader(name string) (FileReader, error)
	Writer() (FileWriter, error)
	List() ([]string, error)
}

type FileWriter interface {
	io.WriteCloser
	Name() string
	Rename(name string) error
	Sync() error
}

type FileReader interface {
	io.ReadSeekCloser
	Name() string
	Size() (int64, error)
}
