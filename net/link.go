package net

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/tasks"
)

// Link is an encrypted communication channel between two identities that can route queries
type Link interface {
	tasks.Runner
	Router
	SetLocalRouter(Router)
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
	Transport() Conn
	Close() error
	Done() <-chan struct{}
}
