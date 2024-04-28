package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

type Module struct {
	config Config
	node   node.Node
	db     *gorm.DB
	log    *log.Logger
	assets resources.Resources

	content content.Module
	storage storage.Module

	indexer *IndexerService

	images   *ImageIndexer
	audio    *AudioIndexer
	matroska *MatroskaIndexer

	indexers map[string]Indexer
}

type Indexer interface {
	content.Describer
	content.Finder
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		mod.indexer,
	).Run(ctx)
}

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	info, err := mod.content.Identify(dataID)
	if err != nil {
		return nil
	}

	if indexer, ok := mod.indexers[info.Type]; ok {
		return indexer.Describe(ctx, dataID, opts)
	}

	return nil
}

func (mod *Module) Find(ctx context.Context, query string, opts *content.FindOpts) (matches []content.Match, err error) {
	if s, _ := mod.audio.Find(ctx, query, opts); len(s) > 0 {
		matches = append(matches, s...)
	}
	if s, _ := mod.images.Find(ctx, query, opts); len(s) > 0 {
		matches = append(matches, s...)
	}
	if s, _ := mod.matroska.Find(ctx, query, opts); len(s) > 0 {
		matches = append(matches, s...)
	}
	return
}
