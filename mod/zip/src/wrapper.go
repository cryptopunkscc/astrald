package zip

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
)

type wrapper struct {
	zipID data.ID
	sets.Set
}

func (mod *Module) wrapper(set sets.Set) (sets.Set, error) {
	var row dbZip
	var err = mod.db.
		Where("set_name = ?", set.Name()).
		First(&row).Error
	if err != nil {
		return set, nil
	}

	return &wrapper{Set: set, zipID: row.DataID}, nil
}

func (w *wrapper) DisplayName() string {
	return fmt.Sprintf("Contents of: {{%s}}", w.zipID)
}
