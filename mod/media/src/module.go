package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/resources"
	"gorm.io/gorm"
	"slices"
)

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
		mod.DescribeObject(ctx, event.ObjectID, astral.DefaultScope())
		return nil
	})

	for event := range mod.Content.Scan(ctx, nil) {
		scope := astral.DefaultScope()

		if slices.Contains(mod.config.AutoIndexNet, event.Type) {
			scope.Zone |= astral.ZoneNetwork
		}

		mod.DescribeObject(ctx, event.ObjectID, scope)
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

func (mod *Module) String() string {
	return media.ModuleName
}
