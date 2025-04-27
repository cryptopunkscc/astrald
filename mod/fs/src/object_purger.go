package fs

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"os"
	"path/filepath"
)

func (mod *Module) PurgeObject(objectID *object.ID, opts *objects.PurgeOpts) (count int, err error) {
	var id = objectID.String()

	for _, dir := range mod.config.Repos {
		var path = filepath.Join(dir, id)

		if os.Remove(path) == nil {
			count++
		}
	}

	return
}
