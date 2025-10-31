package nat

import "github.com/cryptopunkscc/astrald/astral"

type PairPool interface {
	Add(pair *EndpointPair) error
	Take(peer *astral.Identity, local *astral.Identity, isPinger bool) *EndpointPair
	Size() int
}
