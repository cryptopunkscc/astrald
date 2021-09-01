package request

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
)

type Context struct {
	sio.ReadWriteCloser
	Caller    api.Identity
	Port      string
	Observers map[sio.ReadWriteCloser]struct{}
}
