package shares

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ shares.Module = &Module{}

type Module struct {
	config      Config
	node        node.Node
	log         *log.Logger
	assets      assets.Assets
	db          *gorm.DB
	authorizers sig.Set[shares.Authorizer]
	storage     storage.Module
	index       index.Module
}

const localSharePrefix = "mod.share.local."
const setSuffix = ".set"
const publicIndexName = "mod.share.public"

func (mod *Module) Run(ctx context.Context) error {
	tasks.Group(
		NewReadService(mod),
		NewSyncService(mod),
	).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) makeIndexFor(identity id.Identity) error {
	unionName := localSharePrefix + identity.PublicKeyHex()

	info, err := mod.index.IndexInfo(unionName)
	if err == nil {
		if info.Type != index.TypeUnion {
			return fmt.Errorf("index %s is not a union", unionName)
		}
		return nil
	}

	var setName = unionName + setSuffix

	_, err = mod.index.CreateIndex(unionName, index.TypeUnion)
	if err != nil {
		return err
	}

	_, err = mod.index.CreateIndex(setName, index.TypeSet)
	if err != nil {
		return err
	}

	err = mod.index.AddToUnion(unionName, publicIndexName)
	if err != nil {
		return err
	}

	return mod.index.AddToUnion(unionName, setName)
}

func (mod *Module) localShareIndexName(identity id.Identity) string {
	return localSharePrefix + identity.PublicKeyHex()
}

func (mod *Module) addToLocalShareIndex(identity id.Identity, dataID data.ID) error {
	var err = mod.makeIndexFor(identity)
	if err != nil {
		return err
	}

	return mod.index.AddToSet(mod.localShareIndexName(identity)+setSuffix, dataID)
}

func (mod *Module) removeFromLocalShareIndex(identity id.Identity, dataID data.ID) error {
	return mod.index.RemoveFromSet(mod.localShareIndexName(identity)+setSuffix, dataID)
}

func (mod *Module) localShareIndexContains(identity id.Identity, dataID data.ID) (bool, error) {
	return mod.index.Contains(mod.localShareIndexName(identity), dataID)
}
