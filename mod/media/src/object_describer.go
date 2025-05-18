package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	return mod.audio.DescribeObject(ctx, objectID, scope)
}

func (mod *AudioIndexer) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID, opts *astral.Scope) (_ <-chan *objects.SourcedObject, err error) {
	ch := make(chan *objects.SourcedObject, 1)
	defer close(ch)

	audio, err := mod.Index(ctx, objectID)
	if err != nil {
		return ch, err
	}

	ch <- &objects.SourcedObject{
		Source: mod.node.Identity(),
		Object: audio,
	}

	return ch, err
}
