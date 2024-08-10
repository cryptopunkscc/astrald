package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) []*objects.SourcedObject {
	info, err := mod.Content.Identify(objectID)
	if err != nil {
		return nil
	}

	if indexer, ok := mod.indexers[info.Type]; ok {
		return indexer.DescribeObject(ctx, objectID, scope)
	}

	return nil
}
