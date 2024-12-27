package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

// Repository is an interface for creating new data objects in storage
type Repository interface {
	// Name returns repository name
	Name() string

	// Create creates an object in the repository. Repo field should be ignored by repositories.
	Create(opts *CreateOpts) (Writer, error)
	
	// Scan returns a channel of objects stored in the repositiry
	Scan() (<-chan *object.ID, error)

	// Free returns available free space in the repository. -1 if unknown.
	Free() int64
}

// Writer is an interface to write the actual data to objects created by Creators.
type Writer interface {
	// Write data to the object
	Write(p []byte) (n int, err error)

	// Commit commits the written data to storage and returns its ID. Closes the Writer.
	Commit() (object.ID, error)

	// Discard the data written so far and close the Writer.
	Discard() error
}

type CreateOpts struct {
	// Optional. Pre-allocate this much storage.
	Alloc int

	// Optional. Identity requesting object creation.
	As *astral.Identity

	// Optional. Create in this specific repo.
	Repo string
}
