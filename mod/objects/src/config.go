package objects

import (
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

const readServiceName = "objects.read"
const describeServiceName = "objects.describe"
const putServiceName = "objects.put"
const searchServiceName = "objects.search"

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
