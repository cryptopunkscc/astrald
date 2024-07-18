package astral

import "github.com/cryptopunkscc/astrald/id"

// Conn defines the basic interface of an astral connection
type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
}
