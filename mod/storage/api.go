package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"io"
)

type Store interface {
	List() ([]Item, error)
	Open(id data.ID, pos int) (io.ReadCloser, error)
	Stat(id data.ID) Item
	Delete(id data.ID) error
}

type Item struct {
	data.ID
	Availability
}

type Availability int

const (
	StatusUnavailable = Availability(iota)
	StatusAvailable
)
