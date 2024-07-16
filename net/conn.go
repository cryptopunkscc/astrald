package net

import "github.com/cryptopunkscc/astrald/id"

type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
}
