package shares

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/sets"
)

type RemoteShareWrapper struct {
	sets.Set
	row *dbRemoteShare
}

func (mod *Module) remoteShareWrapper(set sets.Set) (sets.Set, error) {
	var row dbRemoteShare
	var err = mod.db.
		Where("set_name = ?", set.Name()).
		First(&row).Error
	if err != nil {
		return set, nil
	}

	return &RemoteShareWrapper{Set: set, row: &row}, nil
}

func (w *RemoteShareWrapper) DisplayName() string {
	return fmt.Sprintf("Remote: {{%s}}@{{%s}}", w.row.Caller, w.row.Target)
}
