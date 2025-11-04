package scheduler

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "scheduler"

// Module is the public interface other modules depend on.
// It intentionally exposes only scheduling capability.
type Module interface {
	Schedule(ctx *astral.Context, action Action)
}
