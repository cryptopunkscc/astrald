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

const methodDescribeAudio = "audio.describe"

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	targetID, err := mod.resolveTarget()
	if err != nil {
		return nil, err
	}

	ch, err := libastrald.WithTarget(targetID).QueryChannel(ctx, methodDescribeAudio, query.Args{
		"id": objectID,
	})
	if err != nil {
		return nil, err
	}

	results := make(chan *objects.Descriptor)

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
			mod.log.Errorv(1, "forward %s to %v: %v", methodDescribeAudio, targetID, err)
		}
	}()

	return results, nil
}
