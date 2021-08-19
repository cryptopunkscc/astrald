package storage

import (
	"io"
)

type Storage interface {
	ReadWriteStorage
	ReadMapStorage
}

type ReadWriteStorage interface {
	ReadStorage
	WriteStorage
}

type ReadMapStorage interface {
	ReadStorage
	MapStorage
}

type ReadStorage interface {
	Reader(name string) (FileReader, error)
	List() ([]string, error)
}

type WriteStorage interface {
	Writer() (FileWriter, error)
}

type MapStorage interface {
	Mapper() (FileMapper, error)
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

type FileMapper interface {
	Map(path string) error
	Rename(name string) error
}