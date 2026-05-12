package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	ch := make(chan *objects.Descriptor, 1)
	defer close(ch)

	row, err := mod.db.FindAudio(objectID)
	if err != nil {
		return ch, err
	}

	ch <- &objects.Descriptor{
		SourceID: mod.node.Identity(),
		ObjectID: objectID,
		Data:     row.ToAudioFile(),
	}

	return ch, err
}
