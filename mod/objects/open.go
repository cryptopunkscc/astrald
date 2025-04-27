package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

func DefaultOpenOpts() *OpenOpts {
	return &OpenOpts{}
}

// Opener is an interface opening data from storage
type Opener interface {
	OpenObject(ctx *astral.Context, objectID object.ID, opts *OpenOpts) (Reader, error)
}

type OpenOpts struct {
	// Open the data at an offset
	Offset uint64

	// Allow opening only from identities accepted by the filter
	QueryFilter astral.IdentityFilter
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
