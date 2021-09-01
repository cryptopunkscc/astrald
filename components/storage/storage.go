package storage

import (
	"io"
)

type ReadWriteStorage interface {
	ReadStorage
	WriteStorage
}

type ReadStorage interface {
	Reader(name string) (FileReader, error)
	List() ([]string, error)
}

type WriteStorage interface {
	Writer() (FileWriter, error)
}

type FileWriter interface {
	io.WriteCloser
	Rename(name string) error
	Sync() error
}

type FileReader interface {
	io.ReadSeekCloser
	Size() (int64, error)
}
