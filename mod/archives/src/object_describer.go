package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (desc []*objects.SourcedObject) {
	if !scope.Zone.Is(astral.ZoneDevice) {
		return
	}

	desc = append(desc, mod.describeArchive(objectID)...)

	return
}

func (mod *Module) describeArchive(objectID object.ID) (list []*objects.SourcedObject) {
	var archive = mod.getCache(objectID)
	if archive == nil {
		return nil
	}

	var totalSize uint64
	for _, e := range archive.Entries {
		totalSize += e.ObjectID.Size
	}

	list = append(list, &objects.SourcedObject{
		Source: mod.node.Identity(),
		Object: &archives.ArchiveDescriptor{
			Format:    archive.Format,
			Entries:   uint32(len(archive.Entries)),
			TotalSize: totalSize,
		},
	})

	return
}
