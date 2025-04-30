package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"gorm.io/gorm"
	"sync"
)

const zipMimeType = "application/zip"

var _ archives.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	db     *gorm.DB

	mu            sync.Mutex
	autoIndexZone astral.Zone
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.autoIndexZone = astral.Zones(mod.config.AutoIndexZones)

	return nil
}

func (mod *Module) String() string {
	return archives.ModuleName
}
