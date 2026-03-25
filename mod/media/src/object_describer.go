package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	return mod.audio.DescribeObject(ctx, objectID)
}

func (mod *AudioIndexer) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (_ <-chan *objects.Descriptor, err error) {
	ch := make(chan *objects.Descriptor, 1)
	defer close(ch)

	audio, err := mod.Index(ctx, objectID)
	if err != nil {
		return ch, err
	}

	ch <- &objects.Descriptor{
		SourceID: mod.node.Identity(),
		ObjectID: objectID,
		Data:     audio,
	}

	return ch, err
}
