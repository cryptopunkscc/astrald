package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/media"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		config: defaultConfig,
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(media.ModuleName, &mod.config)

	mod.db = &DB{assets.Database()}

	err = mod.db.AutoMigrate(&dbAudio{}, &dbObject{})
	if err != nil {
		return nil, err
	}

	mod.audio = NewAudioIndexer(mod)

	mod.ops.AddStruct(mod, "Op")

	return mod, err
}

func init() {
	if err := core.RegisterModule(media.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
