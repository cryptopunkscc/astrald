package archives

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	return mod.describeArchive(objectID)
}

func (mod *Module) describeArchive(objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	var archive = mod.getCache(objectID)
	if archive == nil {
		return nil, errors.New("description unavailable")
	}

	var totalSize uint64
	for _, e := range archive.Entries {
		totalSize += e.ObjectID.Size
	}

	var results = make(chan *objects.DescribeResult, 1)
	defer close(results)

	results <- &objects.DescribeResult{
		OriginID: mod.node.Identity(),
		Descriptor: &archives.ArchiveDescriptor{
			Format:    astral.String8(archive.Format),
			Entries:   astral.Uint32(len(archive.Entries)),
			TotalSize: astral.Uint64(totalSize),
		},
	}

	return results, nil
}
