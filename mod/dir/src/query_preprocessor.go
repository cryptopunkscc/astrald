package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) PreprocessQuery(m *core.QueryModifier) error {
	q := m.Query()

	// never block queries to the local node
	if q.Target.IsEqual(mod.node.Identity()) {
		return nil
	}

	var filters = mod.DefaultFilters()

	// copy filters from the query if provided
	if value, ok := q.Extra.Get("filters"); ok {
		if list, ok := value.([]string); ok && len(list) > 0 {
			filters = list
		}
	}

	// do nothing if there are no filters to apply
	if len(filters) == 0 {
		return nil
	}

	// apply the filters
	if !mod.ApplyFilters(q.Target, filters...) {
		return astral.ErrTargetNotAllowed
	}

	return nil
}
