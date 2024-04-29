package proto

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

type ReadAll struct {
	ID      data.ID `json:"id"`
	Offset  uint64  `json:"offset,omitempty"`
	Virtual bool    `json:"virtual,omitempty"`
	Network bool    `json:"network,omitempty"`
	Filter  string  `json:"filter,omitempty"`
}

type Put struct {
	Bytes []byte `json:"bytes"`
	*storage.CreateOpts
}

type Open struct {
	ID      data.ID `json:"id"`
	Offset  uint64  `json:"offset,omitempty"`
	Virtual bool    `json:"virtual,omitempty"`
	Network bool    `json:"network,omitempty"`
	Filter  string  `json:"filter,omitempty"`
}

type Create struct {
	*storage.CreateOpts
}

type Purge struct {
	ID data.ID `json:"id"`
	*storage.PurgeOpts
}
