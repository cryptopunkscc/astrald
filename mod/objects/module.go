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
	AddOpener(opener Opener, priority int) error
	Opener

	// AddCreator registers a Creator. Creators are queried from highest to lowest priority.
	AddCreator(creator Creator, priority int) error
	Creator

	AddDescriber(Describer) error
	Describer

	AddPurger(purger Purger) error
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

type Describer interface {
	Describe(ctx context.Context, object object.ID, opts *desc.Opts) []*desc.Desc
}

type Purger interface {
	Purge(object.ID, *PurgeOpts) (int, error)
}

type PurgeOpts struct {
	// for future use
}
