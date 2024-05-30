package objects

import (
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

const (
	methodPut      = "objects.put"
	methodRead     = "objects.read"
	methodDescribe = "objects.describe"
	methodRelease  = "objects.release"
	methodSearch   = "objects.search"
	methodHold     = "objects.hold"
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
		relay.CertDesc{}.Type(),
	},
}
