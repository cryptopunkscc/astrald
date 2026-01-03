package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	rows, err := mod.db.FindByObjectID(objectID)
	if err != nil {
		return nil, err
	}

	var results = make(chan *objects.DescribeResult, len(rows))
	defer close(results)

	for _, row := range rows {
		results <- &objects.DescribeResult{
			OriginID: mod.node.Identity(),
			Descriptor: &fs.FileLocation{
				NodeID: mod.node.Identity(),
				Path:   astral.String16(row.Path),
			},
		}
	}

	return results, nil
}
