package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
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
	objects objects.Module

	indexer *IndexerService

	images   *ImageIndexer
	audio    *AudioIndexer
	matroska *MatroskaIndexer

	indexers map[string]Indexer
}

type Indexer interface {
	objects.Describer
	objects.Finder
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		mod.indexer,
	).Run(ctx)
}

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	info, err := mod.content.Identify(objectID)
	if err != nil {
		return nil
	}

	if indexer, ok := mod.indexers[info.Type]; ok {
		return indexer.Describe(ctx, objectID, opts)
	}

	return nil
}

func (mod *Module) Find(ctx context.Context, query string, opts *objects.FindOpts) (matches []objects.Match, err error) {
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
