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
	MethodListPairs         = "nat.list_pairs"
	MethodSetEnabled        = "nat.set_enabled"
)

func init() {
	_ = astral.Add(&PunchSignal{})
	_ = astral.Add(&PairTakeSignal{})
}
