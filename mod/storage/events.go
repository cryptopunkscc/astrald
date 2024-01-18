package storage

import "github.com/cryptopunkscc/astrald/data"

type EventDataCommitted struct {
	DataID data.ID
}

type EventReaderAdded struct {
	Name string
	Reader
}

type EventReaderRemoved struct {
	Name string
	Reader
}

type EventStoreAdded struct {
	Name string
	Store
}

type EventStoreRemoved struct {
	Name string
	Store
}
