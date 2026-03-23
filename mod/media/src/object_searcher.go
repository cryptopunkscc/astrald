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

	if !query.RequiredTagsIn("artist", "album", "title", "genre", "year") {
		return nil, objects.ErrTagNotSupported
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		b := newAudioQuery(mod.db.DB)
		for _, tag := range query.Tags {
			b.Tag(tag)
		}

		rows, err := b.Text(string(query.Query)).Find()
		if err != nil {
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
