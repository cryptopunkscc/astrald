package media

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	info, err := mod.Content.Identify(objectID)
	if err != nil {
		return nil, err
	}

	if indexer, ok := mod.indexers[string(info.Type)]; ok {
		return indexer.DescribeObject(ctx, objectID, scope)
	}

	return nil, errors.New("description unavailable")
}
