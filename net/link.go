package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/tasks"
)

// Link is an encrypted communication channel between two identities that can route queries
type Link interface {
	tasks.Runner
	Router
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
	Transport() SecureConn
	Close() error
	Done() <-chan struct{}
}

// Network returns link's network name or unknown if network could not be determined
func Network(link Link) string {
	var t = link.Transport()
	if t == nil {
		return "unknown"
	}

	if e := t.RemoteEndpoint(); e != nil {
		return e.Network()
	}
	if e := t.LocalEndpoint(); e != nil {
		return e.Network()
	}

	return "unknown"
}
