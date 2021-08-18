package request

import (
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/repo"
	"github.com/cryptopunkscc/astrald/components/serialize"
)

type Context struct {
	Port string
	serialize.Serializer
	repo.ReadWriteRepository
	Observers map[api.Stream]struct{}
}