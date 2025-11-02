package nat

import "github.com/cryptopunkscc/astrald/astral"

type PairPool interface {
	Add(pair *EndpointPair, local *astral.Identity, isPinger bool) error
	Take(peer *astral.Identity) *EndpointPair
	Size() int
}
