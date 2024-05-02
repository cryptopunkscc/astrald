package objects

import (
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/zip"
)

const readServiceName = "objects.read"
const describeServiceName = "objects.describe"

type Config struct {
	DescriptorWhitelist []string
}

var defaultConfig = Config{
	DescriptorWhitelist: []string{
		content.TypeDesc{}.Type(),
		keys.KeyDesc{}.Type(),
		(&media.Audio{}).Type(),
		(&media.Video{}).Type(),
		(&media.Image{}).Type(),
		zip.ArchiveDesc{}.Type(),
		relay.CertDesc{}.Type(),
	},
}
