package shares

import (
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"time"
)

type Config struct {
	NotifyDelay         time.Duration
	DescriptorWhitelist []string
}

var defaultConfig = Config{
	NotifyDelay: 10 * time.Second,
	DescriptorWhitelist: []string{
		content.TypeDesc{}.Type(),
		keys.KeyDesc{}.Type(),
		(&media.Audio{}).Type(),
		archives.ArchiveDesc{}.Type(),
		relay.CertDesc{}.Type(),
	},
}
