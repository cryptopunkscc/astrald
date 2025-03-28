package archives

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if !scope.Zone.Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	return mod.describeArchive(objectID)
}

func (mod *Module) describeArchive(objectID object.ID) (<-chan *objects.SourcedObject, error) {
	var archive = mod.getCache(objectID)
	if archive == nil {
		return nil, errors.New("description unavailable")
	}

	var totalSize uint64
	for _, e := range archive.Entries {
		totalSize += e.ObjectID.Size
	}

	var results = make(chan *objects.SourcedObject, 1)
	defer close(results)

	results <- &objects.SourcedObject{
		Source: mod.node.Identity(),
		Object: &archives.ArchiveDescriptor{
			Format:    archive.Format,
			Entries:   uint32(len(archive.Entries)),
			TotalSize: totalSize,
		},
	}

	return results, nil
}
