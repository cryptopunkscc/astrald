package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/media"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(media.ModuleName, &mod.config)

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbAudio{})
	if err != nil {
		return nil, err
	}

	mod.audio = NewAudioIndexer(mod)

	mod.indexers = map[string]Indexer{
		"audio/mpeg": mod.audio,
		"audio/flac": mod.audio,
		"audio/ogg":  mod.audio,
		"audio/aac":  mod.audio,
		"audio/mp4":  mod.audio,
	}

	return mod, err
}

func init() {
	if err := core.RegisterModule(media.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
