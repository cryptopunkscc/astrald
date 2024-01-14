package storage

import "github.com/cryptopunkscc/astrald/data"

type EventDataAdded DataInfo

type EventDataRemoved struct {
	ID data.ID
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

type EventIndexAdded struct {
	Name string
	Index
}

type EventIndexRemoved struct {
	Name string
	Index
}
