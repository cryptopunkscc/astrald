package request

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/serialize"
)

type Context struct {
	serialize.Serializer
	Port string
	Observers map[api.Stream]struct{}
}
