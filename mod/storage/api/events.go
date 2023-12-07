package storage

import "github.com/cryptopunkscc/astrald/data"

type EventDataAdded DataInfo

type EventDataRemoved struct {
	ID data.ID
}

type EventIndexerAdded struct {
	Indexer
}

type EventIndexerRemoved struct {
	Indexer
}
