package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
)

const ModuleName = "objects"

type Module interface {
	// AddOpener registers an Opener. Openers are queried from highest to lowest priority.
	// Priority ranges:
	// 40-49 - memory cache
	// 30-39 - persistent storage on locally attached devices
	// 20-29 - virtual storage like compressed or generated data
	// 10-19 - storage accessed via network
	AddOpener(name string, opener Opener, priority int) error
	Opener

	// AddCreator registers a Creator. Creators are queried from highest to lowest priority.
	// See AddOpener for priority ranges.
	AddCreator(name string, creator Creator, priority int) error
	Creator

	AddDescriber(Describer) error
	Describer

	AddPurger(name string, purger Purger) error
	Purger

	AddFinder(Finder) error
	Finder

	AddPrototypes(protos ...desc.Data) error
	UnmarshalDescriptor(name string, buf []byte) desc.Data

	// Get reads the whole object into memory and returns the buffer
	Get(id object.ID, opts *OpenOpts) ([]byte, error)

	// Put commits the object to storage and returns its ID
	Put(object []byte, opts *CreateOpts) (object.ID, error)

	Connect(caller id.Identity, target id.Identity) (Consumer, error)
}

type Consumer interface {
	Describe(context.Context, object.ID, *desc.Opts) ([]desc.Data, error)
	Open(context.Context, object.ID, *OpenOpts) (net.SecureConn, error)
}

// Opener is an interface opening data from storage
type Opener interface {
	Open(objectID object.ID, opts *OpenOpts) (Reader, error)
}

// Creator is an interface for creating new data objects in storage
type Creator interface {
	Create(opts *CreateOpts) (Writer, error)
}

type Describer desc.Describer[object.ID]

type Finder interface {
	Find(ctx context.Context, query string, opts *FindOpts) ([]Match, error)
}

type FindOpts struct {
	Network bool
	Virtual bool
	Filter  id.Filter
}

type Match struct {
	ObjectID object.ID
	Score    int
	Exp      string
}

type Purger interface {
	Purge(object.ID, *PurgeOpts) (int, error)
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
	Commit() (object.ID, error)

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
