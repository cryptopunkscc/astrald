package auth

import "github.com/cryptopunkscc/astrald/astral"

// Constraint is any astral.Object that can restrict the scope of a Permit.
// Each action type defines which Constraint types it understands and implements
// ApplyConstraints to evaluate them. Unknown constraint types are handled by
// the action itself (fail-closed or ignored — action decides).
type Constraint = astral.Object
