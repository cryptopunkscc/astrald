package request

import (
	"github.com/cryptopunkscc/astrald/components/sio"
)

type Context struct {
	Port string
	sio.ReadWriteCloser
	Observers map[sio.ReadWriteCloser]struct{}
}
