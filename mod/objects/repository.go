package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

// Repository is an interface for creating new data objects in storage
type Repository interface {
	// Label returns repository label
	Label() string

	// Create creates an object in the repository. Repo field should be ignored by repositories.
	Create(ctx *astral.Context, opts *CreateOpts) (Writer, error)

	// Contains checks if the repository contains the specified object.
	Contains(ctx *astral.Context, objectID *object.ID) (bool, error)

	// Scan returns a channel of objects stored in the repository
	Scan(ctx *astral.Context, follow bool) (<-chan *object.ID, error)

	// Delete deletes an object
	Delete(ctx *astral.Context, objectID *object.ID) error

	// Read reads raw object data
	Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (Reader, error)

	// Free returns available free space in the repository. -1 if unknown.
	Free(ctx *astral.Context) (int64, error)
}

// Writer is an interface to write the actual data to objects created by Creators.
type Writer interface {
	// Write data to the object
	Write(p []byte) (n int, err error)

	// Commit commits the written data to storage and returns its ID. Closes the Writer.
	Commit() (*object.ID, error)

	// Discard the data written so far and close the Writer.
	Discard() error
}

type CreateOpts struct {
	// Optional. Pre-allocate this much storage.
	Alloc int

	// Optional. Create in this specific repo.
	Repo string
}
