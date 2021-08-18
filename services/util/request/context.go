package request

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serializer"
)

type Context struct {
	Port string
	serializer.ReadWriteCloser
	Observers map[api.Stream]struct{}
}
