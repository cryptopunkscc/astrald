package route

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"sync"
)

type replaceIdentity struct {
	net.SecureConn
	sync.Mutex
	remoteIdentity id.Identity
}

func (r *replaceIdentity) Write(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()

	return r.SecureConn.Write(p)
}

func (r *replaceIdentity) Identity() id.Identity {
	return r.remoteIdentity
}
