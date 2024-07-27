package objects

import (
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
)

const (
	methodPut      = "objects.put"
	methodRead     = "objects.read"
	methodDescribe = "objects.describe"
	methodSearch   = "objects.search"
	methodPush     = "objects.push"
)

type Config struct {
	DescriptorWhitelist []string
}

var defaultConfig = Config{
	DescriptorWhitelist: []string{
		content.TypeDesc{}.Type(),
		keys.KeyDesc{}.Type(),
		(&media.Audio{}).Type(),
		archives.ArchiveDesc{}.Type(),
	},
}
