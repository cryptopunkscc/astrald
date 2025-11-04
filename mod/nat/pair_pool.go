package nat

import "github.com/cryptopunkscc/astrald/astral"

type PairPool interface {
	Add(pair *EndpointPair, local *astral.Identity, isPinger bool) error

	// Take returns an idle pair that matches the given peer identity and
	// performs coordinated lock with the peer.
	Take(ctx *astral.Context, peer *astral.Identity) (pair *EndpointPair, err error)

	Size() int
}
