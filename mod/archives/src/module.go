package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
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

func (mod *Module) Run(ctx context.Context) error {
	mod.autoIndexZone = astral.Zones(mod.config.AutoIndexZones)

	for event := range mod.Content.Scan(ctx, &content.ScanOpts{Type: zipMimeType}) {
		mod.Index(ctx, event.ObjectID, &objects.OpenOpts{Zone: mod.autoIndexZone})
	}

	return nil
}

func (mod *Module) String() string {
	return archives.ModuleName
}
