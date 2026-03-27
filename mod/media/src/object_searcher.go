package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) SearchObject(ctx *astral.Context, query objects.SearchQuery) (<-chan *objects.SearchResult, error) {
	return mod.audio.SearchObject(ctx, query)
}

func (mod *AudioIndexer) SearchObject(ctx *astral.Context, query objects.SearchQuery) (<-chan *objects.SearchResult, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	err := query.RequiredTagsIn(knownAudioTags...)
	if err != nil {
		return nil, err
	}

	aq := parseAudioQuery(query)

	results := make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var rows []*dbAudio
		if err := aq.apply(mod.db.Model(&dbAudio{})).Find(&rows).Error; err != nil {
			mod.log.Error("search: db: %v", err)
			return
		}

		for _, row := range rows {
			results <- &objects.SearchResult{
				SourceID: mod.node.Identity(),
				ObjectID: row.ObjectID,
			}
		}
	}()

	return results, nil
}
