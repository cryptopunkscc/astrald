package request

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/sio"
)

type Context struct {
	Port string
	sio.ReadWriteCloser
	Observers map[api.Stream]struct{}
}
