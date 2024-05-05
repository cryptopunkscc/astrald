package objects

import "github.com/cryptopunkscc/astrald/object"

// Creator is an interface for creating new data objects in storage
type Creator interface {
	Create(opts *CreateOpts) (Writer, error)
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
	// Creator should expect the data to be at least Alloc bytes in size
	Alloc int
}
