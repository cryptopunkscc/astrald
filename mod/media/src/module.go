package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/resources"
	"gorm.io/gorm"
	"slices"
)

type Deps struct {
	Admin   admin.Module
	Auth    auth.Module
	Content content.Module
	Objects objects.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	db     *gorm.DB
	log    *log.Logger
	assets resources.Resources

	audio *AudioIndexer

	indexers map[string]Indexer
}

type Indexer interface {
	objects.Describer
	objects.Searcher
}

func (mod *Module) Run(ctx context.Context) error {
	go events.Handle(ctx, mod.node.Events(), func(event objects.EventDiscovered) error {
		mod.Describe(ctx, event.ObjectID, &desc.Opts{})
		return nil
	})

	for event := range mod.Content.Scan(ctx, nil) {
		opts := desc.DefaultOpts()

		if slices.Contains(mod.config.AutoIndexNet, event.Type) {
			opts.Zone |= astral.ZoneNetwork
		}

		mod.Describe(ctx, event.ObjectID, opts)
	}

	return nil
}

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	info, err := mod.Content.Identify(objectID)
	if err != nil {
		return nil
	}

	if indexer, ok := mod.indexers[info.Type]; ok {
		return indexer.Describe(ctx, objectID, opts)
	}

	return nil
}

func (mod *Module) Search(ctx context.Context, query string, opts *objects.SearchOpts) (matches []objects.Match, err error) {
	if s, _ := mod.audio.Search(ctx, query, opts); len(s) > 0 {
		matches = append(matches, s...)
	}
	return
}

func (mod *Module) getParentID(objectID object.ID) (parentID object.ID) {
	mod.db.
		Model(&dbAudio{}).
		Where("picture_id = ?", objectID).
		Select("object_id").
		First(&parentID)

	return
}
