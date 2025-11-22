package nat

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "nat"

// Module defines the NAT traversal module public API.
// Keep minimal for now; NAT orchestration will be implemented in src/.
type Module interface{}

// NOTE: in objects this lies in src/config.go
// but i believe that other modules could want to refer to it (
// and they can only import this package

const (
	MethodStartNatTraversal = "nat.start_traversal"
	MethodPairTake          = "nat.pair_take"
)

func init() {
	_ = astral.DefaultBlueprints.Add(&PunchSignal{})
	_ = astral.DefaultBlueprints.Add(&PairTakeSignal{})
}
