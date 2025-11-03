package nat

import "github.com/cryptopunkscc/astrald/astral"

type PairPool interface {
	Add(pair *EndpointPair, local *astral.Identity, isPinger bool) error
	Take(ctx *astral.Context, peer *astral.Identity) (pair *EndpointPair, err error)
	Size() int
}
