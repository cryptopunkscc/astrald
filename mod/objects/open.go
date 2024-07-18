package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

func DefaultOpenOpts() *OpenOpts {
	return &OpenOpts{
		Zone: astral.DefaultZones,
	}
}

// Opener is an interface opening data from storage
type Opener interface {
	Open(ctx context.Context, objectID object.ID, opts *OpenOpts) (Reader, error)
}

type OpenOpts struct {
	// Limit query to these zones (0 = DefaultZones)
	astral.Zone

	// Open the data at an offset
	Offset uint64

	// Allow opening only from identities accepted by the filter
	QueryFilter id.Filter
}

// Reader is an interface for reading data objects
type Reader interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
	Info() *ReaderInfo
}

type ReaderInfo struct {
	Name string
}
