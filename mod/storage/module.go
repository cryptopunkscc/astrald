package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

const ModuleName = "storage"

type Module interface {
	Opener
	Creator
	Purger

	// ReadAll reads the whole object into memory and returns the buffer
	ReadAll(id data.ID, opts *OpenOpts) ([]byte, error)

	// Put commits the buffer to storage and returns its ID
	Put(bytes []byte, opts *CreateOpts) (data.ID, error)

	// AddOpener registers an Opener. Openers are queried from highest to lowest priority.
	// Priority ranges:
	// 40-49 - memory cache
	// 30-39 - persistent storage on locally attached devices
	// 20-29 - virtual storage like compressed or generated data
	// 10-19 - storage accessed via network
	AddOpener(name string, opener Opener, priority int) error

	// RemoveOpener removes a registered Opener
	RemoveOpener(name string) error

	// AddCreator registers a Creator. Creators are queried from highest to lowest priority.
	// See AddOpener for priority ranges.
	AddCreator(name string, creator Creator, priority int) error

	// RemoveCreator removes a registered Creator
	RemoveCreator(name string) error

	AddPurger(name string, purger Purger) error

	RemovePurger(name string) error
}

// Opener is an interface opening data from storage
type Opener interface {
	Open(dataID data.ID, opts *OpenOpts) (Reader, error)
}

// Creator is an interface for creating new data objects in storage
type Creator interface {
	Create(opts *CreateOpts) (Writer, error)
}

type Purger interface {
	Purge(dataID data.ID, opts *PurgeOpts) (int, error)
}

type PurgeOpts struct {
	// for future use
}

// Reader is an interface for reading data objects
type Reader interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
	Info() *ReaderInfo
}

// Writer is an interface to write the actual data to objects created by Creators.
type Writer interface {
	// Write data to the object
	Write(p []byte) (n int, err error)

	// Commit commits the written data to storage and returns its ID. Closes the Writer.
	Commit() (data.ID, error)

	// Discard the data written so far and close the Writer.
	Discard() error
}

type ReaderInfo struct {
	Name string
}

type OpenOpts struct {
	// Open the data at an offset
	Offset uint64

	// Allow opening from virtual sources (like zip files)
	Virtual bool

	// Allow opening from network sources
	Network bool

	// Allow opening only from identities accepted by the filter
	IdentityFilter id.Filter
}

type CreateOpts struct {
	// Creator should expect the data to be at least Alloc bytes in size
	Alloc int
}
