package media

import (
	"context"
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

const methodSearchAudio = "audio.search"

var _ objects.Searcher = &Module{}

func (mod *Module) SearchObject(ctx *astral.Context, searchQuery objects.SearchQuery) (<-chan *objects.SearchResult, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	targetID, err := mod.resolveTarget()
	if err != nil {
		return nil, err
	}

	ch, err := libastrald.WithTarget(targetID).QueryChannel(ctx, methodSearchAudio, query.Args{
		"q": searchQuery,
	})
	if err != nil {
		return nil, err
	}

	results := make(chan *objects.SearchResult)

	go func() {
		defer close(results)
		defer ch.Close()

		err := ch.Switch(
			channel.Chan(results),
			channel.BreakOnEOS,
			channel.PassErrors,
			channel.WithContext(ctx),
		)
		if err != nil && !errors.Is(err, context.Canceled) {
			mod.log.Errorv(1, "forward %s to %v: %v", methodSearchAudio, targetID, err)
		}
	}()

	return results, nil
}
