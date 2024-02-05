package storage

import "github.com/cryptopunkscc/astrald/data"

type EventDataCommitted struct {
	DataID data.ID
}

type EventOpenerAdded struct {
	Name string
	Opener
}

type EventReaderRemoved struct {
	Name string
	Opener
}

type EventStoreAdded struct {
	Name string
	Creator
}

type EventStoreRemoved struct {
	Name string
	Creator
}
