package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/resources"
	"gorm.io/gorm"
	"slices"
)

type Module struct {
	config Config
	node   node2.Node
	db     *gorm.DB
	log    *log.Logger
	assets resources.Resources

	content content.Module
	objects objects.Module

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

	for event := range mod.content.Scan(ctx, nil) {
		opts := desc.DefaultOpts()

		if slices.Contains(mod.config.AutoIndexNet, event.Type) {
			opts.Zone |= net.ZoneNetwork
		}

		mod.Describe(ctx, event.ObjectID, opts)
	}

	return nil
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

func (mod *Module) Search(ctx context.Context, query string, opts *objects.SearchOpts) (matches []objects.Match, err error) {
	if s, _ := mod.audio.Search(ctx, query, opts); len(s) > 0 {
		matches = append(matches, s...)
	}
	return
}
